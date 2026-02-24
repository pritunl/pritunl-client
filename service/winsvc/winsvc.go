package winsvc

import (
	"github.com/dropbox/godropbox/errors"
	"github.com/judwhite/go-svc"
	"github.com/pritunl/pritunl-client/service/errortypes"
)

type Service struct {
	quit chan bool
}

func (s *Service) Init(env svc.Environment) (err error) {
	return
}

func (s *Service) Start() (err error) {
	return
}

func (s *Service) Stop() (err error) {
	return
}

func (s *Service) Run() (err error) {
	err = svc.Run(s)
	if err != nil {
		err = &errortypes.ExecError{
			errors.Wrap(err, "winsvc: Failed to run service"),
		}
		return
	}

	return
}

func New() (service *Service) {
	service = &Service{
		quit: make(chan bool),
	}

	return
}
