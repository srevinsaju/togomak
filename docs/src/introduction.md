# Togomak

Welcome to Togomak, your go-to solution for seamless Continuous Integration and Continuous Deployment (CICD) workflows. Togomak is an open-source, Golang-based CICD system tool that aims to make building and deploying your applications a breeze, no matter where you want to run them.

## What is Togomak?
Togomak is designed to be an abstraction layer for multiple CICD systems, making it easy to run your builds anywhere. Whether you're developing on your local machine or using popular cloud-based platforms like Jenkins or Cloud Build, Togomak has got you covered.

Taking inspiration from Terraform's lifecycle-based system, Togomak offers a structured approach to managing your build processes. Think of it as GNU Make, but with the added power of working across different environments.

`togomak` doesn't aim to be a competitor to other CI/CD systems like GitHub Actions or Jenkins, but in fact extend on them, helping to create a unified place to track builds across all infrastructure, and to make local developers' build and deployment process much easier. Think of it like this: you are an engineer, and having to deal with pipelines spread across GitLab, GitHub or Jenkins is such a pain. And sometimes, you just want to give a shot checking out the fix for a small typo you made, and now you have to wait for an eternity for the CICD pipelines to be green before you can actually test it on your system, because sometimes, development environments get too complicated, and the Cloud would be the ONLY place where the code actually works. Togomak, tries to unify them.


## Get Started 
Ready to simplify your CICD workflows? Follow our detailed [installation guide](./installation.md) to get Togomak up and running. Explore our documentation to learn more about its [features](./features.md), [configuration options](./configuration.md), and [best practices](./best-practices.md).


## Key Features
### Flexibility and Extensibility 
Togomak supports IPC-based plugins, allowing you to enhance its functionality using Golang or Python. With this capability, you can customize and extend Togomak to suit your specific needs, making it a truly versatile tool for your CICD build systems. Leveraging the power of `hashicorp/go-plugin`, the same plugin system used by Terraform, Togomak offers extensive configuration capabilities for your CICD build systems. Customize and fine-tune your workflows with ease using this robust and battle-tested plugin system.


### Declarative Configuration 
Forget complex configuration setups! Togomak uses a declarative YAML file to define your builds. This makes it straightforward to specify your requirements, ensuring a smooth and hassle-free configuration process. You can now even write your own parsers in any of your favorite languages to write a complex CICD pipeline now. 

### Local Runner System 
Togomak's local runner system is designed to run anywhere. You can detach it and execute your builds in your preferred environment, providing the flexibility and convenience you need as a developer. Say goodbye to limitations and run your builds wherever you choose.

### Remote Tracking System 

> NOTE: this section is planned, and is on the roadmap. See [Roadmap](./roadmap.md)

Collaboration made easy! Togomak includes a remote tracking server that securely stores your build logs and triggers builds through a user-friendly UI. Now you can share your build progress and collaborate with ease, whether you're working locally or on the cloud.

## What's with the name? 
Togomak is a playful fusion of "tokamak" and "to go and make." "Tokamak" refers to a magnetic plasma holding device used in fusion reactors, symbolizing Togomak's aim to provide a solid framework for managing CICD processes. The implementation of Togomak in Golang represents the "to go and make" aspect, leveraging Golang's simplicity, performance, and concurrency to deliver a powerful and efficient solution for your CICD requirements. Together, these elements embody Togomak's commitment to stability, adaptability, and productivity in the world of CICD.

### Okay, that's a lot of talk. How do you pronounce it-
toh-goh-mak. 

For the geeky users out there, Togomak is pronounced as "/toʊˈɡoʊˌmæk/."

## Contribute
Togomak is an open-source project and welcomes contributions from the community. If you have ideas for new features, bug fixes, or improvements, we'd love to hear from you! Check out our [contribution guidelines](./contributing.md) to learn how you can get involved.



