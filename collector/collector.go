package collector

import (
	"log"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/JoseCarlosGarcia95/prometheus-port-exporter/models"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// instancesFile is a string that represents the path to the instances file
	instancesFile string
	// globalLabels is a slice that represents the global labels
	globalLabels []string
	// portOpenedMetric is a prometheus metric that represents the port opened metric
	portOpenedMetric = &prometheus.GaugeVec{}
)

// StartCollector is a function that starts the collector
func StartCollector(_instancesFile string) {
	instancesFile = _instancesFile

	log.Printf("Calculating global labels...\n")
	calculateGlobalLabels()

	log.Printf("Global labels = %v\n", globalLabels)

	log.Printf("Starting collector...\n")
	initializeMetrics()
	go collect()
}

// initializeMetrics is a function that initializes the metrics
func initializeMetrics() {
	portOpenedMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "port_opened",
		Help: "Port opened metric",
	}, globalLabels)

	prometheus.MustRegister(portOpenedMetric)
}

// collect is a function that collects the metrics
func collect() {
	for {
		instances, err := models.ReadInstances(instancesFile)

		if err != nil {
			panic(err)
		}

		for _, instance := range instances {
			log.Printf("Scanning ports in %s...\n", instance.IP)
			ports := PortRange(instance.IP, "tcp", 1, 65535, 1000)

			labels := calculateLabels(instance, 0)

			// Remove all metrics for given labels
			portOpenedMetric.DeletePartialMatch(labels)

			for _, port := range ports {
				labels := calculateLabels(instance, port)
				portOpenedMetric.With(labels).Set(1)
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// IsPortOpen checks if a port is open or not.
func IsPortOpen(host, protocol string, port uint32) bool {
	conn, err := net.DialTimeout(protocol, host+":"+strconv.Itoa(int(port)), 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// PortRange returns a slice of ports that are open.
func PortRange(host, protocol string, start, end, workers uint32) []uint32 {
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var ports = make(chan uint32, end-start+1)

	var openedPorts []uint32

	for i := start; i <= end; i++ {
		ports <- i
	}

	toProcess := make(chan uint32, end-start+1)
	for i := uint32(0); i < end-start+1; i++ {
		toProcess <- i
	}
	close(toProcess)

	for i := uint32(0); i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for port := range toProcess {
				port += start
				if IsPortOpen(host, protocol, port) {
					mutex.Lock()
					openedPorts = append(openedPorts, port)
					mutex.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	close(ports)

	return openedPorts
}

// calculateLabels is a function that calculates the labels
func calculateLabels(instance *models.Instance, port uint32) prometheus.Labels {
	labels := make(prometheus.Labels)

	for _, label := range globalLabels {
		labels[label] = instance.Labels[label]
	}

	labels["ip"] = instance.IP

	if port != 0 {
		labels["port"] = strconv.Itoa(int(port))
	}

	return labels
}

// calculateGlobalLabels is a function that calculates the global labels
func calculateGlobalLabels() {
	_globalLabels := make(map[string]bool)

	instances, err := models.ReadInstances(instancesFile)

	if err != nil {
		panic(err)
	}

	for _, instance := range instances {
		for label := range instance.Labels {
			_globalLabels[label] = true
		}
	}

	for label := range _globalLabels {
		globalLabels = append(globalLabels, label)
	}

	globalLabels = append(globalLabels, "ip")
	globalLabels = append(globalLabels, "port")
}
