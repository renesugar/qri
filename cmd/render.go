package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/qri-io/qri/core"
	"github.com/spf13/cobra"
)

// NewRenderCommand creates a new `qri render` command for executing templates against datasets
func NewRenderCommand(f Factory, ioStreams IOStreams) *cobra.Command {
	o := &RenderOptions{IOStreams: ioStreams}
	cmd := &cobra.Command{
		Use:   "render",
		Short: "execute a template against a dataset",
		Long:  `the most common use for render is to generate html from a qri dataset`,
		Example: `  render a dataset called me/schools:
  $ qri render -o=schools.html me/schools

  render a dataset with a custom template:
  $ qri render --template=template.html me/schools`,
		Annotations: map[string]string{
			"group": "dataset",
		},
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ExitIfErr(o.Complete(f, args))
			ExitIfErr(o.Run())
		},
	}

	cmd.Flags().StringVarP(&o.Template, "template", "t", "", "path to template file")
	cmd.Flags().StringVarP(&o.Output, "output", "o", "", "path to write output file")
	cmd.Flags().BoolVarP(&o.All, "all", "a", false, "read all dataset entries (overrides limit, offest)")
	cmd.Flags().IntVarP(&o.Limit, "limit", "l", 50, "max number of records to read")
	cmd.Flags().IntVarP(&o.Offset, "offset", "s", 0, "number of records to skip")

	return cmd
}

// RenderOptions encapsulates state for the render command
type RenderOptions struct {
	IOStreams

	Ref      string
	Template string
	Output   string
	All      bool
	Limit    int
	Offset   int

	RenderRequests *core.RenderRequests
}

// Complete adds any missing configuration that can only be added just before calling Run
func (o *RenderOptions) Complete(f Factory, args []string) (err error) {
	o.Ref = args[0]
	o.RenderRequests, err = f.RenderRequests()
	return
}

// Run executes the render command
func (o *RenderOptions) Run() (err error) {
	var template []byte

	if o.Template != "" {
		template, err = ioutil.ReadFile(o.Template)
		if err != nil {
			return err
		}
	}

	p := &core.RenderParams{
		Ref:            o.Ref,
		Template:       template,
		TemplateFormat: "html",
		All:            o.All,
		Limit:          o.Limit,
		Offset:         o.Offset,
	}

	res := []byte{}
	if err = o.RenderRequests.Render(p, &res); err != nil {
		return err
	}

	if o.Output == "" {
		fmt.Fprint(o.Out, string(res))
	} else {
		ioutil.WriteFile(o.Output, res, 0777)
	}
	return nil
}
