package danmep

import (
    "fmt"
    "log"
    "net"
    "os/exec"

    danmtypes "github.com/nokia/danm/crd/apis/danm/v1"
)

type calicoReleaseIPServiceImpl struct {}

func (h *calicoReleaseIPServiceImpl) IsIPAllocatedByMe(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) bool {
    return ep.Spec.NetworkType == "calico"
}

func (h *calicoReleaseIPServiceImpl) ReleaseIP(dnet *danmtypes.DanmNet, ep *danmtypes.DanmEp, ip string) error {
    parsedIp := net.ParseIP(ip)
    if parsedIp == nil {
        parsedIp, _, _ = net.ParseCIDR(ip)
    }
    cmd := exec.Command("calicoctl", "ipam", "release", fmt.Sprintf("--ip=%s", parsedIp))
    log.Println("release calico managed IP:" + cmd.String())

    if output, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("could not release calico managed IP %s, because: %s | output: %s", ip, err, output)
    }
    return nil
}
