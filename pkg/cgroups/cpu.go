package cgroups

import (
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
)

//CPUAcct cpuacct control group controller
type CPUAcct struct {
	path    string
	cgroup2 bool
}

//NewCPUAcct create new CPUAcct type
func NewCPUAcct(path string) (*CPUAcct, error) {
	cpuacct := &CPUAcct{
		path: path,
	}

	var err error
	cpuacct.cgroup2, err = IsCgroup2UnifiedMode()
	if err != nil {
		return nil, errors.Wrap(err, "determining cgroup version")
	}
	return cpuacct, nil
}

//Stats get cpuacct stats
func (ca *CPUAcct) Stats() ([]byte, error) {
	if ca.cgroup2 {
		return ca.statsV2()
	}
	return ca.statsV1()
}

func (ca *CPUAcct) statsV1() ([]byte, error) {
	stats, err := ioutil.ReadFile(filepath.Join(ca.path, "cpuacct.usage"))
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving cpu stats cgroup v1")
	}
	return stats, nil
}

func (ca *CPUAcct) statsV2() ([]byte, error) {
	stats, err := ioutil.ReadFile(filepath.Join(ca.path, "cpu.stat"))
	if err != nil {
		return nil, errors.Wrapf(err, "retrieving cpu stats cgroup v2")
	}
	return stats, nil
}

// GetSystemCPUUsage returns the system usage for all the cgroups
func GetSystemCPUUsage() (uint64, error) {
	cgroupv2, err := IsCgroup2UnifiedMode()
	if err != nil {
		return 0, err
	}
	if !cgroupv2 {
		p := filepath.Join(cgroupRoot, "CPUAcct", "cpuacct.usage")
		return readFileAsUint64(p)
	}

	files, err := ioutil.ReadDir(cgroupRoot)
	if err != nil {
		return 0, errors.Wrapf(err, "read directory %q", cgroupRoot)
	}
	var total uint64
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		p := filepath.Join(cgroupRoot, file.Name(), "cpu.stat")

		values, err := readCgroup2MapPath(p)
		if err != nil {
			return 0, err
		}

		if val, found := values["usage_usec"]; found {
			v, err := strconv.ParseUint(cleanString(val[0]), 10, 0)
			if err != nil {
				return 0, err
			}
			total += v * 1000
		}

	}
	return total, nil
}