package main

// https://nsl.cz/using-tun-tap-in-go-or-how-to-write-vpn/

// To run it just exec on one host something like this:
// ./tun -local 10.168.2.11/24 -remote {remote ipv4}

// and on the second:
// ./tun -local 10.168.2.9/24 -remote {remote ipv4}

import (
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/exec"
    "golang.org/x/net/ipv4"
    "github.com/songgao/water"
)


const (
    // I use TUN interface, so only plain IP packet, no ethernet header + mtu is set to 1300
    BUFFERSIZE = 1500
    MTU        = "1300"
)



var (
    localIP  = flag.String("local", "", "Local tun interface IP/MASK like 192.168.3.3/24")
    remoteIP = flag.String("remote", "", "Remote server (external) IP like 8.8.8.8")
    port     = flag.Int("port", 4321, "UDP port for communication")
)



func runIP(args ...string) {
    cmd := exec.Command("/sbin/ip", args...)
    cmd.Stderr = os.Stderr
    cmd.Stdout = os.Stdout
    cmd.Stdin = os.Stdin
    err := cmd.Run()
    if nil != err {
        log.Fatalln("Error running /sbin/ip:", err)
    }
}

func main() {
    flag.Parse()
    // check if we have anything
    if "" == *localIP {
        flag.Usage()
        log.Fatalln("\nlocal ip is not specified")
    }
    if "" == *remoteIP {
        flag.Usage()
        log.Fatalln("\nremote server is not specified")
    }
    // create TUN interface
    // iface, err := water.NewTUN("")

    iface, err := water.New(water.Config{
      DeviceType: water.TUN,
      })

    if nil != err {
        log.Fatalln("Unable to allocate TUN interface:", err)
    }
    log.Println("Interface allocated:", iface.Name())
    // set interface parameters
    runIP("link", "set", "dev", iface.Name(), "mtu", MTU)
    runIP("addr", "add", *localIP, "dev", iface.Name())
    runIP("link", "set", "dev", iface.Name(), "up")
    // reslove remote addr
    remoteAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%v", *remoteIP, *port))
    if nil != err {
        log.Fatalln("Unable to resolve remote addr:", err)
    }
    // listen to local socket
    lstnAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%v", *port))
    if nil != err {
        log.Fatalln("Unable to get UDP socket:", err)
    }
    lstnConn, err := net.ListenUDP("udp", lstnAddr)
    if nil != err {
        log.Fatalln("Unable to listen on UDP socket:", err)
    }
    defer lstnConn.Close()
    // recv in separate thread
    go func() {
        buf := make([]byte, BUFFERSIZE)
        for {
            n, addr, err := lstnConn.ReadFromUDP(buf)
            // just debug
            header, _ := ipv4.ParseHeader(buf[:n])
            fmt.Printf("Received %d bytes from %v: %+v\n", n, addr, header)
            if err != nil || n == 0 {
                fmt.Println("Error: ", err)
                continue
            }
            // write to TUN interface
            iface.Write(buf[:n])
        }
    }()
    // and one more loop
    packet := make([]byte, BUFFERSIZE)
    for {
        plen, err := iface.Read(packet)
        if err != nil {
            break
        }
        // debug :)
        header, _ := ipv4.ParseHeader(packet[:plen])
        fmt.Printf("Sending to remote: %+v (%+v)\n", header, err)
        // real send
        lstnConn.WriteToUDP(packet[:plen], remoteAddr)
    }
}
