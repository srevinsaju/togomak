package data

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/internal/conductor"
	"github.com/srevinsaju/togomak/v1/internal/global"
	"github.com/srevinsaju/togomak/v1/internal/ui"
	"github.com/srevinsaju/togomak/v1/internal/x"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
)

type TfProviderConfig struct {
	source     string
	allowApply bool
	vars       map[string]cty.Value
}

const (
	TfBlockArgumentSource     = "source"
	TfBlockArgumentAllowApply = "allow_apply"
	TfBlockArgumentVars       = "vars"
)

type TfProvider struct {
	initialized bool

	ctx context.Context
	cfg TfProviderConfig
}

func (e *TfProvider) Logger() *logrus.Entry {
	return global.Logger().WithField("provider", e.Name())
}

func (e *TfProvider) Name() string {
	return "tf"
}

func (e *TfProvider) Identifier() string {
	return "data.tf"
}

func (e *TfProvider) SetContext(context context.Context) {
	e.ctx = context
}

func (e *TfProvider) Version() string {
	return "1"
}

func (e *TfProvider) Url() string {
	return "embedded::togomak.srev.in/providers/data/tf"
}

func (e *TfProvider) DecodeBody(conductor conductor.Conductor, body hcl.Body, opts ...ProviderOption) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics
	hclContext := conductor.Eval().Context()

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = diags.Extend(d)

	conductor.Eval().Mutex().RLock()
	source, d := content.Attributes[TfBlockArgumentSource].Expr.Value(hclContext)
	conductor.Eval().Mutex().RUnlock()

	diags = diags.Extend(d)

	var allowApply cty.Value
	var vars cty.Value

	attr, ok := content.Attributes[TfBlockArgumentAllowApply]
	if ok {
		conductor.Eval().Mutex().RLock()
		allowApply, d = attr.Expr.Value(hclContext)
		conductor.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
	}
	attr, ok = content.Attributes[TfBlockArgumentVars]
	if ok {
		conductor.Eval().Mutex().RLock()
		vars, d = attr.Expr.Value(hclContext)
		conductor.Eval().Mutex().RUnlock()
		diags = diags.Extend(d)
	}
	var varsGo map[string]cty.Value
	if !vars.IsNull() {
		varsGo = vars.AsValueMap()
	}

	e.cfg = TfProviderConfig{
		source:     source.AsString(),
		allowApply: !allowApply.IsNull() && allowApply.True(),
		vars:       varsGo,
	}

	return diags
}

func (e *TfProvider) New() Provider {
	return &TfProvider{
		initialized: true,
	}
}

func (e *TfProvider) Schema() *hcl.BodySchema {
	return &hcl.BodySchema{
		Attributes: []hcl.AttributeSchema{
			{
				Name:     TfBlockArgumentSource,
				Required: true,
			},
			{
				Name:     TfBlockArgumentAllowApply,
				Required: false,
			},
			{
				Name:     TfBlockArgumentVars,
				Required: false,
			},
		},
	}
}

func (e *TfProvider) Initialized() bool {
	return e.initialized
}

func (e *TfProvider) Value(conductor conductor.Conductor, ctx context.Context, id string, opts ...ProviderOption) (string, hcl.Diagnostics) {
	if !e.initialized {
		panic("provider not initialized")
	}
	return "", nil
}

func (e *TfProvider) Attributes(conductor conductor.Conductor, ctx context.Context, id string, opts ...ProviderOption) (map[string]cty.Value, hcl.Diagnostics) {
	logger := e.Logger()
	tmpDir := global.TempDir()
	cfg := NewProviderConfig(opts...)

	var diags hcl.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}
	var attrs = make(map[string]cty.Value)

	logger.Tracef("downloading %s to %s", e.cfg.source, tmpDir)
	terraformWorkdir := filepath.Join(tmpDir, "data", "terraform", e.Identifier())
	dst := filepath.Join(terraformWorkdir, "src")
	client := getter.Client{
		Ctx: ctx,
		Src: e.cfg.source,
		Dst: dst,
		Dir: true,
		Pwd: cfg.Paths.Cwd,
	}

	logger.Tracef("downloading source")
	ppb := ui.NewPassiveProgressBar(logger, fmt.Sprintf("pulling %s", e.cfg.source))
	ppb.Init()
	err := client.Get()
	ppb.Done()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to retrieve source",
			Detail:   err.Error(),
		})
		return attrs, diags
	}

	logger.Tracef("searching for terraform binary")
	execPath, err := exec.LookPath("terraform")
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "terraform not found",
			Detail:   err.Error(),
		})
		return attrs, diags
	}

	logger.Tracef("terraform found at %s", execPath)
	workingDir := dst
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to start terraform",
			Detail:   err.Error(),
		})
	}

	ppb = ui.NewPassiveProgressBar(logger, fmt.Sprintf("reading %s.%s", e.Identifier(), id))
	ppb.Init()
	defer ppb.Done()
	logger.Tracef("running terraform init on %s", dst)
	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to initialize terraform",
			Detail:   err.Error(),
		})
	}

	logger.Tracef("resolving absolute path to plan file")
	planDir := filepath.Join(terraformWorkdir, "plan")
	x.Must(os.MkdirAll(planDir, 0755))
	planFile, err := filepath.Abs(filepath.Join(planDir, fmt.Sprintf("plan%s.out", uuid.New().String())))
	x.Must(err)
	logger.Tracef("running terraform plan on %s writing plan file to %s", dst, planFile)
	x.Must(os.MkdirAll(filepath.Dir(planFile), 0755))

	var vars []tfexec.PlanOption
	for k, v := range e.cfg.vars {
		logger.Tracef("setting variable %s to %s", k, v.AsString())
		vars = append(vars, tfexec.Var(fmt.Sprintf("%s=%s", k, v.AsString())))
	}
	planOpts := []tfexec.PlanOption{tfexec.Lock(false), tfexec.Out(planFile)}
	planOpts = append(planOpts, vars...)
	infraChanged, err := tf.Plan(ctx, planOpts...)
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to plan terraform",
			Detail:   err.Error(),
		})
	}

	logger.Tracef("infra changes detected: %t", infraChanged)
	logger.Tracef("user allowed apply: %t", e.cfg.allowApply)
	if infraChanged && !e.cfg.allowApply {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "terraform plan has changes",
			Detail:   "infrastructure changes were detected when running terraform plan. make sure you don't have any resources in the same directory as the terraform data block",
		})
	}

	err = tf.Apply(ctx, tfexec.Lock(false), tfexec.DirOrPlan(planFile))
	state, err := tf.Show(ctx)
	if err != nil {
		log.Fatalf("error running Show: %s", err)
	}

	tfMap := map[string]interface{}{}
	for _, m := range state.Values.RootModule.Resources {
		_, ok := tfMap[m.Type]
		if !ok {
			tfMap[m.Type] = map[string]interface{}{}
		}
		tfMap[m.Type].(map[string]interface{})[m.Name] = m.AttributeValues
	}

	for k, v := range tfMap {
		attrs[k] = cty.ObjectVal(getObjectType(v.(map[string]interface{})))
	}
	// get the commit
	return attrs, diags
}

// getObjectType recursively converts a map[string]interface{} to map[string]cty.Value
func getObjectType(m map[string]interface{}) map[string]cty.Value {
	s := map[string]cty.Value{}
	for k, v := range m {
		typeRaw := reflect.TypeOf(v)
		if typeRaw == nil {
			s[k] = cty.StringVal("")
			continue
		}
		typeOf := typeRaw.Kind()
		if typeOf == reflect.Map {
			s[k] = cty.ObjectVal(getObjectType(v.(map[string]interface{})))
		} else if typeOf == reflect.Slice {
			s[k] = cty.ListVal(getListType(v.([]interface{})))
		} else {
			impliedType, err := gocty.ImpliedType(v)
			if err != nil {
				panic(err)
			}
			impliedValue, err := gocty.ToCtyValue(v, impliedType)
			if err != nil {
				panic(err)
			}
			s[k] = impliedValue
		}
	}
	return s
}

// getListType recursively converts a []interface{} to []cty.Value
func getListType(m []interface{}) []cty.Value {
	var s []cty.Value
	for _, v := range m {
		typeOf := reflect.TypeOf(v).Kind()
		if typeOf == reflect.Map {
			s = append(s, cty.ObjectVal(getObjectType(v.(map[string]interface{}))))
		} else if typeOf == reflect.Slice {
			s = append(s, cty.ListVal(getListType(v.([]interface{}))))
		} else {
			impliedType, err := gocty.ImpliedType(v)
			if err != nil {
				panic(err)
			}
			impliedValue, err := gocty.ToCtyValue(v, impliedType)
			if err != nil {
				panic(err)
			}
			s = append(s, impliedValue)
		}
	}
	return s
}
