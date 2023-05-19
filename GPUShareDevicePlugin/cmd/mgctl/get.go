package main

import (
	"fmt"
	pbdevice "github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/gen/proto/go/device/v1"
	"github.com/AccessibleAI/cnvrg-fractional-accelerator-device-plugin/pkg/ctlutils"
	"github.com/atomicgo/cursor"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var getCmd = &cobra.Command{
	Use:     "get",
	Aliases: []string{"g"},
	Short:   "get resources",
}

var processGetParams = []param{
	{name: "watch", shorthand: "w", value: false, usage: "watch for the changes"},
}

var processesGetCmd = &cobra.Command{
	Use:     "processes",
	Aliases: []string{"p", "process"},
	Short:   "list gpu processes and processes metadata",
	Run: func(cmd *cobra.Command, args []string) {
		getDevicesProcesses()
	},
}

var getDevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"d", "device"},
	Short:   "get gpu devices",
	Run: func(cmd *cobra.Command, args []string) {
		getDevices()
	},
}

func getDevices() {
	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString("addr"))
	if conn == nil {
		log.Fatalf("can't initiate connection to metagpu server")
	}
	defer conn.Close()
	device := pbdevice.NewDeviceServiceClient(conn)
	resp, err := device.GetMetaDeviceInfo(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetMetaDeviceInfoRequest{})
	if err != nil {
		log.Fatal(err)
	}
	to := &TableOutput{}
	to.header = table.Row{"Idx", "UUID", "Memory", "Shares", "Share size"}
	to.body, to.footer = buildDeviceInfoTableBody(resp.Devices)
	to.buildTable()
	to.print()

}

func getDevicesProcesses() {

	conn := ctlutils.GetGrpcMetaGpuSrvClientConn(viper.GetString("addr"))
	if conn == nil {
		log.Fatalf("can't initiate connection to metagpu server")
	}
	defer conn.Close()
	device := pbdevice.NewDeviceServiceClient(conn)
	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("faild to detect podId, err: %s", err)
	}

	to := &TableOutput{}
	to.header = table.Row{"Pod", "NS", "Device", "Node", "GPU", "Memory", "Pid", "Cmd", "Req"}

	if viper.GetBool("watch") {
		request := &pbdevice.StreamGpuContainersRequest{PodId: hostname}
		stream, err := device.StreamGpuContainers(ctlutils.AuthenticatedContext(viper.GetString("token")), request)
		if err != nil {
			log.Fatal(err)
		}

		refreshCh := make(chan bool)
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		go func() {
			for {
				time.Sleep(1 * time.Second)
				refreshCh <- true
			}
		}()

		for {
			select {
			case <-sigCh:
				cursor.ClearLine()
				log.Info("shutting down")
				os.Exit(0)
			case <-refreshCh:
				processResp, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					log.Fatalf("error watching gpu processes, err: %s", err)
				}
				deviceResp, err := device.GetDevices(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetDevicesRequest{})
				if err != nil {
					log.Errorf("falid to list devices, err: %s ", err)
					return
				}
				to.body = buildDeviceProcessesTableBody(processResp.GpuContainers)
				to.footer = buildDeviceProcessesTableFooter(processResp.GpuContainers, deviceResp.Device, processResp.VisibilityLevel)
				to.buildTable()
				to.print()
			}
		}
	} else {
		processResp, err := device.GetGpuContainers(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetGpuContainersRequest{PodId: hostname})
		if err != nil {
			log.Errorf("falid to list device processes, err: %s ", err)
			return
		}
		deviceResp, err := device.GetDevices(ctlutils.AuthenticatedContext(viper.GetString("token")), &pbdevice.GetDevicesRequest{})
		if err != nil {
			log.Errorf("falid to list devices, err: %s ", err)
			return
		}
		to.body = buildDeviceProcessesTableBody(processResp.GpuContainers)
		to.footer = buildDeviceProcessesTableFooter(processResp.GpuContainers, deviceResp.Device, processResp.VisibilityLevel)
		to.buildTable()
		to.print()
	}
}

func buildDeviceInfoTableBody(devices []*pbdevice.Device) (body []table.Row, footer table.Row) {
	var totMem uint64
	var shares uint32
	for _, d := range devices {
		shares = d.Shares
		totMem += d.MemoryTotal
		body = append(body, table.Row{
			d.Index,
			d.Uuid,
			d.MemoryTotal,
			d.Shares,
			d.MemoryShareSize,
		})
	}
	footer = table.Row{len(devices), "", fmt.Sprintf("%dMB", totMem), uint32(len(devices)) * shares, ""}
	return body, footer
}

func buildDeviceProcessesTableBody(containers []*pbdevice.GpuContainer) (body []table.Row) {

	for _, c := range containers {
		if len(c.ContainerDevices) > 0 {
			maxMem := int64(c.ContainerDevices[0].Device.MemoryShareSize * uint64(c.MetagpuRequests))
			if len(c.DeviceProcesses) > 0 {
				for _, p := range c.DeviceProcesses {
					relativeGpuUsage := (p.GpuUtilization * 100) / (100 / c.ContainerDevices[0].Device.Shares * uint32(c.MetagpuRequests))
					gpuUsage := fmt.Sprintf("\u001B[32m%d%%\u001B[0m", relativeGpuUsage)
					if relativeGpuUsage > 100 {
						gpuUsage = fmt.Sprintf("\u001B[31m%d%%\u001B[0m", relativeGpuUsage)
					}
					memUsage := fmt.Sprintf("\u001B[32m%d\u001B[0m/%d", p.Memory, maxMem)
					if int64(p.Memory) > maxMem {
						memUsage = fmt.Sprintf("\u001B[31m%d\u001B[0m/%d", p.Memory, maxMem)
					}
					body = append(body, table.Row{
						c.PodId,
						c.PodNamespace,
						formatContainerDeviceIndexes(c),
						c.NodeName,
						gpuUsage,
						memUsage,
						p.Pid,
						p.Cmdline,
						c.MetagpuRequests,
					})
				}
			} else {
				memUsage := fmt.Sprintf("\u001B[32m%d\u001B[0m/%d", 0, maxMem)
				body = append(body, table.Row{
					c.PodId,
					c.PodNamespace,
					formatContainerDeviceIndexes(c),
					c.NodeName,
					"-",
					memUsage,
					"-",
					"-",
					c.MetagpuRequests,
				})
			}
		} else {
			body = append(body, table.Row{
				c.PodId,
				c.PodNamespace,
				formatContainerDeviceIndexes(c),
				c.NodeName,
				"-",
				"-",
				"-",
				"-",
				c.MetagpuRequests,
			})
		}

	}

	return
}

func buildDeviceProcessesTableFooter(containers []*pbdevice.GpuContainer, devices map[string]*pbdevice.Device, vl string) (footer table.Row) {
	metaGpuSummary := fmt.Sprintf("%d", getTotalRequests(containers))
	// TODO: fix this, the vl should be taken from directly form the  package
	// to problem is that package now includes the nvidia linux native stuff
	// and some package re-org is required
	//if vl == "l0" { // TODO: temporary disabled
	metaGpuSummary = fmt.Sprintf("%d/%d", getTotalShares(devices), getTotalRequests(containers))
	//}
	usedMem := fmt.Sprintf("%dMb", getTotalMemoryUsedByProcesses(containers))
	return table.Row{len(containers), "", "", "", "", usedMem, "", "", metaGpuSummary}
}
