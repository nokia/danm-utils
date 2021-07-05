package danmep

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	danmipam "github.com/nokia/danm/pkg/ipam"
)

type calicoReleaseIPServiceImpl releaseIPServiceImplBase

func (h *calicoReleaseIPServiceImpl) IsIPAllocatedByMe(ip string, neType string) bool {
	return ip != danmipam.NoneAllocType && ip != "" &&
		!danmipam.WasIpAllocatedByDanm(ip, h.dnet.Spec.Options.Cidr) &&
		!danmipam.WasIpAllocatedByDanm(ip, h.dnet.Spec.Options.Pool6.Cidr) &&
		neType == "calico"
}

func (h *calicoReleaseIPServiceImpl) ReleaseIP(ip string) error {
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		parsedIp, _, _ = net.ParseCIDR(ip)
	}
	cmd := exec.Command("calicoctl", "ipam", "release", fmt.Sprintf("--ip=%s", parsedIp))
	log.Printf("release calico managed IP: %s", cmd)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("could not release calico managed IP %s, because: %s | output: %s", ip, err, output)
	}
	return nil
}
