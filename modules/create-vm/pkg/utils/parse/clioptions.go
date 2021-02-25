package parse

import (
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/constants"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/create-vm/pkg/utils/output"
	"github.com/kubevirt/kubevirt-tekton-tasks/modules/shared/pkg/zutils"
	"go.uber.org/zap/zapcore"
	"strings"
)

const (
	vmManifestOptionName        = "vm-manifest"
	vmNamespaceOptionName       = "vm-namespace"
	templateNameOptionName      = "template-name"
	templateNamespaceOptionName = "template-namespace"
	templateParamsOptionName    = "template-params"
)

const templateParamSep = ":"

type CLIOptions struct {
	TemplateName              string            `arg:"--template-name,env:TEMPLATE_NAME" placeholder:"NAME" help:"Name of a template to create VM from"`
	TemplateNamespace         string            `arg:"--template-namespace,env:TEMPLATE_NAMESPACE" placeholder:"NAMESPACE" help:"Namespace of a template to create VM from"`
	TemplateParams            []string          `arg:"--template-params" placeholder:"KEY1:VAL1 KEY2:VAL2" help:"Template params to pass when processing the template manifest"`
	VirtualMachineManifest    string            `arg:"--vm-manifest,env:VM_MANIFEST" placeholder:"MANIFEST" help:"YAML manifest of a VirtualMachine resource to be created (can be set by VM_MANIFEST env variable)."`
	VirtualMachineNamespace   string            `arg:"--vm-namespace,env:VM_NAMESPACE" placeholder:"NAMESPACE" help:"Namespace where to create the VM"`
	DataVolumes               []string          `arg:"--dvs" placeholder:"DV1 DV2" help:"Add DataVolumes to VM Volumes"`
	OwnDataVolumes            []string          `arg:"--own-dvs" placeholder:"DV1 DV2" help:"Add DataVolumes to VM Volumes and add VM to DV ownerReferences. These DVs will be deleted once the created VM gets deleted."`
	PersistentVolumeClaims    []string          `arg:"--pvcs" placeholder:"PVC1 PVC2" help:"Add PersistentVolumeClaims to VM Volumes."`
	OwnPersistentVolumeClaims []string          `arg:"--own-pvcs" placeholder:"PVC1 PVC2" help:"Add PersistentVolumeClaims to VM Volumes and add VM to PVC ownerReferences. These PVCs will be deleted once the created VM gets deleted."`
	Output                    output.OutputType `arg:"-o" placeholder:"FORMAT" help:"Output format. One of: yaml|json"`
	Debug                     bool              `arg:"--debug" help:"Sets DEBUG log level"`
}

func (c *CLIOptions) GetAllPVCNames() []string {
	return zutils.ConcatStringSlices(c.OwnPersistentVolumeClaims, c.PersistentVolumeClaims)
}

func (c *CLIOptions) GetAllDVNames() []string {
	return zutils.ConcatStringSlices(c.OwnDataVolumes, c.DataVolumes)
}

func (c *CLIOptions) GetAllDiskNames() []string {
	return zutils.ConcatStringSlices(c.GetAllPVCNames(), c.GetAllDVNames())
}

func (c *CLIOptions) GetTemplateParams() map[string]string {
	result := make(map[string]string, len(c.TemplateParams))

	lastKey := ""

	for _, keyVal := range c.TemplateParams {
		split := strings.SplitN(keyVal, templateParamSep, 2)

		switch len(split) {
		case 1:
			// expect space between values and append to the last key seen
			if lastKey != "" {
				result[lastKey] += " " + split[0]
			}
		case 2:
			lastKey = strings.TrimSpace(split[0])
			result[lastKey] = split[1]
		}
	}
	return result
}

func (c *CLIOptions) GetDebugLevel() zapcore.Level {
	if c.Debug {
		return zapcore.DebugLevel
	}
	return zapcore.InfoLevel
}

func (c *CLIOptions) GetCreationMode() constants.CreationMode {
	if c.VirtualMachineManifest != "" && c.TemplateName != "" {
		return ""
	}
	if c.VirtualMachineManifest != "" {
		return constants.VMManifestCreationMode
	}

	if c.TemplateName != "" {
		return constants.TemplateCreationMode
	}

	return ""
}

func (c *CLIOptions) GetTemplateNamespace() string {
	return c.TemplateNamespace
}

func (c *CLIOptions) GetVirtualMachineManifest() string {
	return c.VirtualMachineManifest
}

func (c *CLIOptions) GetVirtualMachineNamespace() string {
	return c.VirtualMachineNamespace
}

func (c *CLIOptions) Init() error {
	if err := c.assertValidMode(); err != nil {
		return err
	}

	if err := c.assertValidTypes(); err != nil {
		return err
	}

	if err := c.resolveTemplateParams(); err != nil {
		return err
	}

	if err := c.resolveDefaultNamespacesAndManifests(); err != nil {
		return err
	}

	c.trimSpaces()

	return nil
}
