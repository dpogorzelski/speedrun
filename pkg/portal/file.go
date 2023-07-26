package portal

import (
	"context"
	"io/fs"
	"os"

	"github.com/apex/log"
	"github.com/dpogorzelski/speedrun/proto/portal"
)

func (s *Server) FileRead(ctx context.Context, file *portal.FileReadRequest) (*portal.FileReadResponse, error) {
	fields := log.Fields{
		"context": "file",
		"command": "read",
		"name":    file.GetPath(),
	}
	log := log.WithFields(fields)
	log.Debug("Received file read request")

	content, err := os.ReadFile(file.GetPath())
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &portal.FileReadResponse{State: portal.State_UNKNOWN, Content: string(content)}, nil
}

func (s *Server) FileCp(ctx context.Context, file *portal.FileCpRequest) (*portal.FileCpResponse, error) {
	fields := log.Fields{
		"context": "file",
		"command": "cp",
		"name":    file.GetDst(),
	}
	log := log.WithFields(fields)
	log.Debug("Received file cp request")

	if file.GetRemoteSrc() && file.GetRemoteDst() {
		content, err := os.ReadFile(file.GetSrc())
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		err = os.WriteFile(file.GetDst(), content, 0644)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

	} else if file.GetRemoteDst() {
		err := os.WriteFile(file.GetDst(), file.GetContent(), 0644)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	} else if file.GetRemoteSrc() {
		content, err := os.ReadFile(file.GetSrc())
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
		return &portal.FileCpResponse{State: portal.State_UNKNOWN, Content: content}, nil
	}
	return &portal.FileCpResponse{State: portal.State_UNKNOWN}, nil
}

func (s *Server) FileChmod(ctx context.Context, file *portal.FileChmodRequest) (*portal.FileChmodResponse, error) {
	fields := log.Fields{
		"context": "file",
		"command": "read",
		"name":    file.GetPath(),
	}
	log := log.WithFields(fields)
	log.Debug("Received file chmod request")

	err := os.Chmod(file.GetPath(), fs.FileMode(file.GetFilemode()))
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	return &portal.FileChmodResponse{State: portal.State_UNKNOWN}, nil
}
