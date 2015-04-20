package Dust2

import (
	"net"
	"os"
	"path/filepath"

	"git.torproject.org/pluggable-transports/goptlib.git"
	"git.torproject.org/pluggable-transports/obfs4.git/transports/base"
	"github.com/blanu/Dust/go/v2/interface"
)

const (
	idFilename = "Dust2_id"
)

type serverFactory struct {
	transport *Transport
	stateDir  string
	private   *Dust.ServerPrivate
}

func (t *Transport) ServerFactory(stateDir string, args *pt.Args) (base.ServerFactory, error) {
	idPath := filepath.Join(stateDir, idFilename)

	unparsed, err := inPtArgs(args)
	if err != nil {
		return nil, err
	}

	var private *Dust.ServerPrivate

	_, statErr := os.Stat(idPath)
	switch {
	case statErr == nil:
		// ID file exists.  Try to load it.
		private, err = Dust.LoadServerPrivateFile(idPath)
		if err != nil {
			log.Error("loading identity file from %s: %s", idPath, err)
			return nil, err
		}

		// TODO: doesn't check ID file for congruence with existing parameters.

	case statErr != nil && os.IsNotExist(statErr):
		// ID file doesn't exist.  Try to write a new one.
		ep, err := Dust.ParseEndpointParams(unparsed)
		if err != nil {
			log.Error("parsing endpoint parameters: %s", err)
			return nil, err
		}

		private, err = Dust.NewServerPrivate(ep)
		if err != nil {
			log.Error("generating new identity: %s", err)
			return nil, err
		}

		err = private.SavePrivateFile(idPath)
		if err != nil {
			log.Error("saving new identity to %s: %s", idPath, err)
			return nil, err
		}
		
	default:
		log.Error("stat %s: %s", idPath, statErr)
		return nil, statErr
	}

	return &serverFactory{
		transport: t,
		stateDir:  stateDir,
		private:   private,
	}, nil
}

func (sf *serverFactory) Transport() base.Transport {
	return sf.transport
}

func (sf *serverFactory) Args() *pt.Args {
	return outPtArgs(sf.private.Public().Unparse())
}

func (sf *serverFactory) WrapConn(visible net.Conn) (net.Conn, error) {
	rconn, err := Dust.BeginRawStreamServer(visible, sf.private)
	if err != nil {
		return nil, err
	}

	return rconn, nil
}