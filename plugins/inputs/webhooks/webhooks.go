//go:generate ../../../tools/readme_config_includer/generator
package webhooks

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/gorilla/mux"

	"Dana"
	"Dana/config"
	"Dana/plugins/inputs"
	"Dana/plugins/inputs/webhooks/artifactory"
	"Dana/plugins/inputs/webhooks/filestack"
	"Dana/plugins/inputs/webhooks/github"
	"Dana/plugins/inputs/webhooks/mandrill"
	"Dana/plugins/inputs/webhooks/papertrail"
	"Dana/plugins/inputs/webhooks/particle"
	"Dana/plugins/inputs/webhooks/rollbar"
)

//go:embed sample.conf
var sampleConfig string

const (
	defaultReadTimeout  = 10 * time.Second
	defaultWriteTimeout = 10 * time.Second
)

type Webhook interface {
	Register(router *mux.Router, acc Dana.Accumulator, log Dana.Logger)
}

func init() {
	inputs.Add("webhooks", func() Dana.Input { return NewWebhooks() })
}

type Webhooks struct {
	ServiceAddress string          `toml:"service_address"`
	ReadTimeout    config.Duration `toml:"read_timeout"`
	WriteTimeout   config.Duration `toml:"write_timeout"`

	Github      *github.GithubWebhook           `toml:"github"`
	Filestack   *filestack.FilestackWebhook     `toml:"filestack"`
	Mandrill    *mandrill.MandrillWebhook       `toml:"mandrill"`
	Rollbar     *rollbar.RollbarWebhook         `toml:"rollbar"`
	Papertrail  *papertrail.PapertrailWebhook   `toml:"papertrail"`
	Particle    *particle.ParticleWebhook       `toml:"particle"`
	Artifactory *artifactory.ArtifactoryWebhook `toml:"artifactory"`

	Log Dana.Logger `toml:"-"`

	srv *http.Server
}

func NewWebhooks() *Webhooks {
	return &Webhooks{}
}

func (*Webhooks) SampleConfig() string {
	return sampleConfig
}

func (*Webhooks) Gather(Dana.Accumulator) error {
	return nil
}

// AvailableWebhooks Looks for fields which implement Webhook interface
func (wb *Webhooks) AvailableWebhooks() []Webhook {
	webhooks := make([]Webhook, 0)
	s := reflect.ValueOf(wb).Elem()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		if !f.CanInterface() {
			continue
		}

		if wbPlugin, ok := f.Interface().(Webhook); ok {
			if !reflect.ValueOf(wbPlugin).IsNil() {
				webhooks = append(webhooks, wbPlugin)
			}
		}
	}

	return webhooks
}

func (wb *Webhooks) Start(acc Dana.Accumulator) error {
	if wb.ReadTimeout < config.Duration(time.Second) {
		wb.ReadTimeout = config.Duration(defaultReadTimeout)
	}
	if wb.WriteTimeout < config.Duration(time.Second) {
		wb.WriteTimeout = config.Duration(defaultWriteTimeout)
	}

	r := mux.NewRouter()

	for _, webhook := range wb.AvailableWebhooks() {
		webhook.Register(r, acc, wb.Log)
	}

	wb.srv = &http.Server{
		Handler:      r,
		ReadTimeout:  time.Duration(wb.ReadTimeout),
		WriteTimeout: time.Duration(wb.WriteTimeout),
	}

	ln, err := net.Listen("tcp", wb.ServiceAddress)
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	go func() {
		if err := wb.srv.Serve(ln); err != nil {
			if err != http.ErrServerClosed {
				acc.AddError(fmt.Errorf("error listening: %w", err))
			}
		}
	}()

	wb.Log.Infof("Started the webhooks service on %s", wb.ServiceAddress)

	return nil
}

func (wb *Webhooks) Stop() {
	wb.srv.Close()
	wb.Log.Infof("Stopping the Webhooks service")
}
