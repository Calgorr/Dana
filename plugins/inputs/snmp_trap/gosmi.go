package snmp_trap

import (
	"Dana"
	"Dana/internal/snmp"
)

type gosmiTranslator struct {
}

func (*gosmiTranslator) lookup(oid string) (snmp.MibEntry, error) {
	return snmp.TrapLookup(oid)
}

func newGosmiTranslator(paths []string, log Dana.Logger) (*gosmiTranslator, error) {
	err := snmp.LoadMibsFromPath(paths, log, &snmp.GosmiMibLoader{})
	if err == nil {
		return &gosmiTranslator{}, nil
	}
	return nil, err
}
