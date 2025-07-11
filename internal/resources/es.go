package resources

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/doors-dev/doors/internal/common"
	"github.com/evanw/esbuild/pkg/api"
)


type BuildErrors []api.Message

func (b BuildErrors) Error() string {
	var errs []error
	for _, m := range b {
		var formatted string
		if m.Location != nil {
			formatted = fmt.Sprintf("%s:%d:%d: %s",
				m.Location.File,
				m.Location.Line,
				m.Location.Column,
				m.Text)
		} else {
			formatted = m.Text
		}
		errs = append(errs, errors.New(formatted))
	}
	return errors.Join(errs...).Error()
}

func buildES(options *api.BuildOptions) ([]byte, error) {
	options.Write = false
	options.Platform = api.PlatformBrowser
	result := api.Build(*options)
	if len(result.Errors) != 0 {
		for _, m := range result.Errors {
			slog.Error("esbuild error", slog.String("text", m.Text))
		}

		return nil, BuildErrors(result.Errors)

	}
	if len(result.OutputFiles) == 0 {
		return nil, BuildErrors([]api.Message{{
			Text: "no output produced",
		}})
	}
	data := result.OutputFiles[0].Contents
	return data, nil
}

func Build(entry string, opt api.BuildOptions) ([]byte, error) {
	opt.EntryPoints = []string{entry}
	return buildES(&opt)
}

func BuildFS(fs fs.FS, entry string, opt api.BuildOptions) ([]byte, error) {
	opt.EntryPoints = []string{entry}
	if opt.Plugins == nil {
		opt.Plugins = []api.Plugin{fsPlugin(fs)}
	} else {
		opt.Plugins = append(opt.Plugins, fsPlugin(fs))
	}
	return buildES(&opt)
}

func Bundle(entry string, o api.BuildOptions) ([]byte, error) {
	o.EntryPoints = []string{entry}
	o.Format = api.FormatESModule
	o.Bundle = true
	return buildES(&o)
}

func BundleFS(fs fs.FS, entry string, o api.BuildOptions) ([]byte, error) {
	o.EntryPoints = []string{entry}
	o.Format = api.FormatESModule
	o.Bundle = true
	if o.Plugins == nil {
		o.Plugins = []api.Plugin{fsPlugin(fs)}
	} else {
		o.Plugins = append(o.Plugins, fsPlugin(fs))
	}
	return buildES(&o)
}

func Transform(path string, o api.BuildOptions) ([]byte, error) {
	o.EntryPoints = []string{path}
	return buildES(&o)
}

func TransformBytes(content []byte, o api.BuildOptions) ([]byte, error) {
	o.Stdin = &api.StdinOptions{
		Contents:   common.AsString(&content),
		Sourcefile: "index.js",
		Loader:     api.LoaderJS,
	}
	return buildES(&o)
}

func TransformBytesTS(content []byte, o api.BuildOptions) ([]byte, error) {
	o.Stdin = &api.StdinOptions{
		Contents:   common.AsString(&content),
		Sourcefile: "index.ts",
		Loader:     api.LoaderTS,
	}
	return buildES(&o)
}
