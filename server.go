package main

import (
	"flag"
	"github.com/miekg/dns"
	"github.com/samalba/dockerclient"
	"log"
	"net"
	"strings"
)

var docker dockerclient.DockerClient
var dockerBr = flag.String("i", "docker0", "docker bridge")

func handleDnsRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	records := make([]dns.RR, 0)
	q := r.Question[0]

	if q.Qtype == dns.TypeA && strings.HasSuffix(q.Name, ".docker.") {
		docker, _ := dockerclient.NewDockerClient("unix:///var/run/docker.sock")
		nameDomain := strings.Split(q.Name, ".")
		containers, err := docker.ListContainers(false)
		if err != nil {
			log.Fatal(err)
		}
		for _, c := range containers {
			for _, n := range c.Names {
				if nameDomain[0] == n[1:] {
					info, _ := docker.InspectContainer(c.Id)
					log.Printf("Container %s has ip %s\n", nameDomain[0], info.NetworkSettings.IpAddress)
					records = append(records,
						&dns.A{Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
							A: net.ParseIP(info.NetworkSettings.IpAddress)})
				}
			}
		}

	}

	m.Answer = append(m.Answer, records...)
	defer w.WriteMsg(m)
}

func main() {
	flag.Parse()
	dockerInt, err := net.InterfaceByName(*dockerBr)
	if err != nil {
		panic(err)
	}

	dockerIpNet, err := dockerInt.Addrs()
	if err != nil {
		panic(err)
	}

	dockerIp := strings.Split(dockerIpNet[0].String(), "/")

	server := &dns.Server{Addr: dockerIp[0] + ":53", Net: "udp"}
	dns.HandleFunc(".", handleDnsRequest)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
