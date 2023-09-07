package data

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/sirupsen/logrus"
	"github.com/srevinsaju/togomak/v1/pkg/c"
	"github.com/srevinsaju/togomak/v1/pkg/ui"
	"github.com/srevinsaju/togomak/v1/pkg/x"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
)

type TfProviderConfig struct {
	source string
}

const (
	TfBlockArgumentSource = "source"
)

type TfProvider struct {
	initialized bool
	Default     hcl.Expression `hcl:"default" json:"default"`

	ctx context.Context
	cfg TfProviderConfig
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

func (e *TfProvider) DecodeBody(body hcl.Body) hcl.Diagnostics {
	if !e.initialized {
		panic("provider not initialized")
	}
	var diags hcl.Diagnostics
	hclContext := e.ctx.Value(c.TogomakContextHclEval).(*hcl.EvalContext)

	schema := e.Schema()
	content, d := body.Content(schema)
	diags = diags.Extend(d)

	source, d := content.Attributes[TfBlockArgumentSource].Expr.Value(hclContext)
	diags = diags.Extend(d)

	e.cfg = TfProviderConfig{
		source: source.AsString(),
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
		},
	}
}

func (e *TfProvider) Initialized() bool {
	return e.initialized
}

func (e *TfProvider) Value(ctx context.Context, id string) (string, hcl.Diagnostics) {
	if !e.initialized {
		panic("provider not initialized")
	}
	return "", nil
}

func (e *TfProvider) Attributes(ctx context.Context, id string) (map[string]cty.Value, hcl.Diagnostics) {
	logger := ctx.Value(c.TogomakContextLogger).(*logrus.Logger).WithField("provider", e.Name())
	tmpDir := ctx.Value(c.TogomakContextTempDir).(string)
	cwd := ctx.Value(c.TogomakContextCwd).(string)

	var diags hcl.Diagnostics
	if !e.initialized {
		panic("provider not initialized")
	}
	var attrs = make(map[string]cty.Value)

	dst := filepath.Join(tmpDir, "data", "terraform", e.Identifier())
	client := getter.Client{
		Ctx:              ctx,
		Src:              e.cfg.source,
		Dst:              dst,
		Dir:              true,
		Pwd:              cwd,
		ProgressListener: &TfProgressBar{logger: logger},
	}
	err := client.Get()
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to retrieve source",
			Detail:   err.Error(),
		})
		return attrs, diags
	}

	execPath, err := exec.LookPath("terraform")
	if err != nil {
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "terraform not found",
			Detail:   err.Error(),
		})
		return attrs, diags
	}

	workingDir := dst
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to start terraform",
			Detail:   err.Error(),
		})
	}

	err = tf.Init(ctx, tfexec.Upgrade(true))
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to initialize terraform",
			Detail:   err.Error(),
		})
	}
	planFile, err := filepath.Abs(filepath.Join(dst, "plan.out"))
	x.Must(err)
	x.Must(os.MkdirAll(filepath.Dir(planFile), 0755))

	infraChanged, err := tf.Plan(
		ctx,
		tfexec.Lock(false),
		tfexec.Out(planFile))
	if err != nil {
		return attrs, diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "failed to plan terraform",
			Detail:   err.Error(),
		})
	}
	if infraChanged {
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
		fmt.Printf("%s.%s\n", m.Type, m.Name)
		_, ok := tfMap[m.Type]
		if !ok {
			tfMap[m.Type] = map[string]interface{}{}
		}
		tfMap[m.Type].(map[string]interface{})[m.Name] = m.AttributeValues
	}

	for k, v := range tfMap {
		attrs[k] = cty.MapVal(getMapType(v.(map[string]interface{})))
	}
	// get the commit
	return attrs, diags
}

type TfProgressBar struct {
	logger *logrus.Entry
	src    string
	pb     *ui.ProgressWriter
}

func (e *TfProgressBar) Init() {
	e.pb = ui.NewProgressWriter(e.logger, fmt.Sprintf("downloading %s", e.src))
}

func (e *TfProgressBar) TrackProgress(src string, currentSize, totalSize int64, stream io.ReadCloser) (body io.ReadCloser) {
	for {
		_, err := io.CopyN(e.pb, stream, 1)
		if err != nil {
			e.pb.Close()
			return stream
		}
	}
}

func getMapType(m map[string]interface{}) map[string]cty.Value {
	s := map[string]cty.Value{}
	for k, v := range m {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			s[k] = cty.MapVal(getMapType(v.(map[string]interface{})))
		} else if reflect.TypeOf(v).Kind() == reflect.Slice {
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
			m[k] = impliedValue
		}
	}
	return s
}
func getListType(m []interface{}) []cty.Value {
	var s []cty.Value
	for _, v := range m {
		if reflect.TypeOf(v).Kind() == reflect.Map {
			s = append(s, cty.MapVal(getMapType(v.(map[string]interface{}))))
		} else if reflect.TypeOf(v).Kind() == reflect.Slice {
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
