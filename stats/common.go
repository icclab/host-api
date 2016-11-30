package stats

import (
	"bufio"
	"encoding/json"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/shirou/gopsutil/mem"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"strings"
	"time"
)

func pathParts(path string) []string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	return strings.Split(path, "/")
}

func parseRequestToken(tokenString string, parsedPublicKey interface{}) (*jwt.Token, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("No JWT token provided")
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return parsedPublicKey, nil
	})
}

func getContainerStats(reader *bufio.Reader, count int, id string, pid int) (containerInfo, error) {
	i, err := getDockerContainerInfo(reader, count, id, pid)
	return i, err
}

type DockerStats struct {
	Read      time.Time `json:"read"`
	PidsStats struct {
		Current int64 `json:"current"`
	} `json:"pids_stats"`
	Networks map[string]struct {
		RxBytes   int64 `json:"rx_bytes"`
		RxDropped int64 `json:"rx_dropped"`
		RxErrors  int64 `json:"rx_errors"`
		RxPackets int64 `json:"rx_packets"`
		TxBytes   int64 `json:"tx_bytes"`
		TxDropped int64 `json:"tx_dropped"`
		TxErrors  int64 `json:"tx_errors"`
		TxPackets int64 `json:"tx_packets"`
	} `json:"networks"`
	BlkioStats struct {
		IoServiceBytesRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_service_bytes_recursive"`
		IoServicedRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_serviced_recursive"`
		IoQueueRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_queue_recursive"`
		IoServiceTimeRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_service_time_recursive"`
		IoWaitTimeRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_wait_time_recursive"`
		IoMergedRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_merged_recursive"`
		IoTimeRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"io_time_recursive"`
		SectorsRecursive []struct {
			Major int64  `json:"major"`
			Minor int64  `json:"minor"`
			Op    string `json:"op"`
			Value int64  `json:"value"`
		} `json:"sectors_recursive"`
	} `json:"blkio_stats"`
	MemoryStats struct {
		Stats struct {
			TotalPgmajfault         int64  `json:"total_pgmajfault"`
			Cache                   int64  `json:"cache"`
			MappedFile              int64  `json:"mapped_file"`
			TotalInactiveFile       int64  `json:"total_inactive_file"`
			Pgpgout                 int64  `json:"pgpgout"`
			Rss                     int64  `json:"rss"`
			TotalMappedFile         int64  `json:"total_mapped_file"`
			Writeback               int64  `json:"writeback"`
			Unevictable             int64  `json:"unevictable"`
			Pgpgin                  int64  `json:"pgpgin"`
			TotalUnevictable        int64  `json:"total_unevictable"`
			Pgmajfault              int64  `json:"pgmajfault"`
			TotalRss                int64  `json:"total_rss"`
			TotalRssHuge            int64  `json:"total_rss_huge"`
			TotalWriteback          int64  `json:"total_writeback"`
			TotalInactiveAnon       int64  `json:"total_inactive_anon"`
			RssHuge                 int64  `json:"rss_huge"`
			HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit"`
			TotalPgfault            int64  `json:"total_pgfault"`
			TotalActiveFile         int64  `json:"total_active_file"`
			ActiveAnon              int64  `json:"active_anon"`
			TotalActiveAnon         int64  `json:"total_active_anon"`
			TotalPgpgout            int64  `json:"total_pgpgout"`
			TotalCache              int64  `json:"total_cache"`
			InactiveAnon            int64  `json:"inactive_anon"`
			ActiveFile              int64  `json:"active_file"`
			Pgfault                 int64  `json:"pgfault"`
			InactiveFile            int64  `json:"inactive_file"`
			TotalPgpgin             int64  `json:"total_pgpgin"`
		} `json:"stats"`
		MaxUsage int64 `json:"max_usage"`
		Usage    int64 `json:"usage"`
		Failcnt  int64 `json:"failcnt"`
		Limit    int64 `json:"limit"`
	} `json:"memory_stats"`
	CPUStats struct {
		CPUUsage struct {
			PercpuUsage       []int64 `json:"percpu_usage"`
			UsageInUsermode   int64   `json:"usage_in_usermode"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelmode int64   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		ThrottlingData struct {
			Periods          int64 `json:"periods"`
			ThrottledPeriods int64 `json:"throttled_periods"`
			ThrottledTime    int64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			PercpuUsage       []int64 `json:"percpu_usage"`
			UsageInUsermode   int64   `json:"usage_in_usermode"`
			TotalUsage        int64   `json:"total_usage"`
			UsageInKernelmode int64   `json:"usage_in_kernelmode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		ThrottlingData struct {
			Periods          int64 `json:"periods"`
			ThrottledPeriods int64 `json:"throttled_periods"`
			ThrottledTime    int64 `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"precpu_stats"`
}

type containerInfo struct {
	Id    string
	Stats []*containerStats
}

type containerStats struct {
	Timestamp time.Time    `json:"timestamp"`
	Cpu       CpuStats     `json:"cpu,omitempty"`
	DiskIo    DiskIoStats  `json:"diskio,omitempty"`
	Network   NetworkStats `json:"network,omitempty"`
	Memory    MemoryStats  `json:"memory,omitempty"`
}

type CpuStats struct {
	Usage CpuUsage `json:"usage"`
}

type CpuUsage struct {
	// Total CPU usage.
	// Units: nanoseconds
	Total uint64 `json:"total"`

	// Per CPU/core usage of the container.
	// Unit: nanoseconds.
	PerCpu []uint64 `json:"per_cpu_usage,omitempty"`

	// Time spent in user space.
	// Unit: nanoseconds
	User uint64 `json:"user"`

	// Time spent in kernel space.
	// Unit: nanoseconds
	System uint64 `json:"system"`
}

type DiskIoStats struct {
	IoServiceBytes []PerDiskStats `json:"io_service_bytes,omitempty"`
}

type PerDiskStats struct {
	Major uint64            `json:"major"`
	Minor uint64            `json:"minor"`
	Stats map[string]uint64 `json:"stats"`
}

type NetworkStats struct {
	InterfaceStats
	Interfaces []InterfaceStats `json:"interfaces,omitempty"`
}

type MemoryStats struct {
	// Current memory usage, this includes all memory regardless of when it was
	// accessed.
	// Units: Bytes.
	Usage uint64 `json:"usage"`
}

type InterfaceStats struct {
	// The name of the interface.
	Name string `json:"name"`
	// Cumulative count of bytes received.
	RxBytes uint64 `json:"rx_bytes"`
	// Cumulative count of packets received.
	RxPackets uint64 `json:"rx_packets"`
	// Cumulative count of receive errors encountered.
	RxErrors uint64 `json:"rx_errors"`
	// Cumulative count of packets dropped while receiving.
	RxDropped uint64 `json:"rx_dropped"`
	// Cumulative count of bytes transmitted.
	TxBytes uint64 `json:"tx_bytes"`
	// Cumulative count of packets transmitted.
	TxPackets uint64 `json:"tx_packets"`
	// Cumulative count of transmit errors encountered.
	TxErrors uint64 `json:"tx_errors"`
	// Cumulative count of packets dropped while transmitting.
	TxDropped uint64 `json:"tx_dropped"`
}

func getContainerInfo(reader *bufio.Reader, count int, id string, pid int) (containerInfo, error) {
	contInfo := containerInfo{}
	contInfo.Id = id
	stats := []*containerStats{}
	for i := 0; i < count; i++ {
		str, err := reader.ReadString([]byte("\n")[0])
		if err != nil {
			return containerInfo{}, err
		}
		dockerStats, err := FromString(str)
		if err != nil {
			return containerInfo{}, err
		}
		contStats := convertDockerStats(dockerStats, pid)
		stats = append(stats, contStats)
	}
	contInfo.Stats = stats
	return contInfo, nil
}

func convertDockerStats(stats DockerStats, pid int) *containerStats {
	containerStats := containerStats{}
	containerStats.Timestamp = stats.Read
	containerStats.Cpu.Usage.Total = uint64(stats.CPUStats.CPUUsage.TotalUsage)
	containerStats.Cpu.Usage.PerCpu = []uint64{}
	for _, value := range stats.CPUStats.CPUUsage.PercpuUsage {
		containerStats.Cpu.Usage.PerCpu = append(containerStats.Cpu.Usage.PerCpu, uint64(value))
	}
	containerStats.Cpu.Usage.System = uint64(stats.CPUStats.CPUUsage.UsageInKernelmode)
	containerStats.Cpu.Usage.User = uint64(stats.CPUStats.CPUUsage.UsageInKernelmode)
	containerStats.Memory.Usage = uint64(stats.MemoryStats.Usage)
	containerStats.Network.Interfaces = []InterfaceStats{}
	for name, netStats := range getLinkStats(pid) {
		data := InterfaceStats{}
		data.Name = name
		data.RxBytes = uint64(netStats.RxBytes)
		data.RxDropped = uint64(netStats.RxDropped)
		data.RxErrors = uint64(netStats.RxErrors)
		data.RxPackets = uint64(netStats.RxPackets)
		data.TxBytes = uint64(netStats.TxBytes)
		data.TxDropped = uint64(netStats.TxDropped)
		data.TxPackets = uint64(netStats.TxPackets)
		data.TxErrors = uint64(netStats.TxErrors)
		containerStats.Network.Interfaces = append(containerStats.Network.Interfaces, data)
	}
	containerStats.DiskIo.IoServiceBytes = []PerDiskStats{}
	for _, diskStats := range stats.BlkioStats.IoServiceBytesRecursive {
		data := PerDiskStats{}
		data.Stats = map[string]uint64{}
		data.Stats[diskStats.Op] = uint64(diskStats.Value)
		containerStats.DiskIo.IoServiceBytes = append(containerStats.DiskIo.IoServiceBytes, data)
	}
	return &containerStats
}

func FromString(rawstring string) (DockerStats, error) {
	obj := DockerStats{}
	err := json.Unmarshal([]byte(rawstring), &obj)
	if err != nil {
		return obj, err
	}
	return obj, nil
}

func getMemCapcity() (uint64, error) {
	data, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return data.Total, nil
}

func getLinkStats(pid int) map[string]*netlink.LinkStatistics {
	ret := map[string]*netlink.LinkStatistics{}
	nsHandler, err := netns.GetFromPid(pid)
	if err != nil {
		return nil
	}
	defer nsHandler.Close()
	handler, err := netlink.NewHandleAt(nsHandler)
	if err != nil {
		return nil
	}
	defer handler.Delete()
	links, err := handler.LinkList()
	if err != nil {
		return nil
	}
	for _, link := range links {
		attr := link.Attrs()
		if attr.Name != "lo" {
			ret[attr.Name] = attr.Statistics
		}
	}
	return ret
}
