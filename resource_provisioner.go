package main

import (
	"github.com/hashicorp/terraform/terraform"
	"github.com/mitchellh/mapstructure"
	"log"
	"time"
	"runtime"
	"errors"
)

type ResourceProvisioner struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Script   string `mapstructure:"script"`
	Params   string `mapstructure:"params"`
}

func (r *ResourceProvisioner) Apply(
	o terraform.UIOutput,
	s *terraform.InstanceState,
	c *terraform.ResourceConfig) error {

	provisioner, err := r.decodeConfig(c)
	if err != nil {
		o.Output("erred out here")
		return err
	}

	err = provisioner.Validate()
	if err != nil {
		o.Output("Invalid provisioner configuration settings")
		return err
	}

	// ensure that this is a windows machine
	if runtime.GOOS != "windows" {
		o.Output("psinvoke is only supported on Windows at this time.")
		return errors.New("psinvoke is only supported on Windows at this time.")
	}

	err = provisioner.Run(o)
	if err != nil {
		o.Output("erred out here 4")
		return err
	}

	return nil
}

func (r *ResourceProvisioner) Validate(c *terraform.ResourceConfig) (ws []string, es []error) {
	provisioner, err := r.decodeConfig(c)
	if err != nil {
		es = append(es, err)
		return ws, es
	}

	err = provisioner.Validate()
	if err != nil {
		es = append(es, err)
		return ws, es
	}

	return ws, es
}

func (r *ResourceProvisioner) decodeConfig(c *terraform.ResourceConfig) (*Provisioner, error) {
	// decodes configuration from terraform and builds out a provisioner
	p := new(Provisioner)
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           p,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, err
	}

	// build a map of all configuration values, by default this is going to
	// pass in all configuration elements for the base configuration as
	// well as extra values. Build a single value and then from there, continue forth!
	m := make(map[string]interface{})
	for k, v := range c.Raw {
		m[k] = v
	}
	for k, v := range c.Config {
		m[k] = v
	}

	err = decoder.Decode(m)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func retryFunc(timeout time.Duration, f func() error) error {
	finish := time.After(timeout)

	for {
		err := f()
		if err == nil {
			return nil
		}
		log.Printf("Retryable error: %v", err)

		select {
		case <-finish:
			return err
		case <-time.After(3 * time.Second):
		}
	}
}
