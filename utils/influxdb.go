package utils

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/influxdata/influxdb/client/v2"

	"node/common"
)

var conf *common.Settings

func init() {
	conf = common.GetSettings()
}

func GetInfluxDBWriteClient() (client.Client, error) {
	//conf := common.GetSettings()
	addr := conf.Getv("INFLUXDB_HOST") + ":" + conf.Getv("INFLUXDB_WRITE_PORT")
	return getInfluxDBClient(addr)
}

func GetInfluxDBREADClient() (client.Client, error) {
	//conf := common.GetSettings()
	addr := conf.Getv("INFLUXDB_HOST") + ":" + conf.Getv("INFLUXDB_WRITE_PORT")
	return getInfluxDBClient(addr)
}

// defer client.Close() when getclient
func getInfluxDBClient(addr string) (client.Client, error) {
	client, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: addr,
	})
	return client, err
}

func newBatchPoints() (client.BatchPoints, error) {
	//conf := common.GetSettings()
	database := conf.Getv("INFLUXDB_DATABASE")
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: database,
	})
	return bp, err
}

func WriteData(cli client.Client, measurement string, tags map[string]string, fields map[string]interface{}, t time.Time) (bool, error) {
	// Create a new point batch
	batchpoints, err := newBatchPoints()
	if err != nil {
		logrus.Errorf("influxdb new batchpoints error: %v", err)
		return false, err
	}
	// Create a point and add to batch
	logrus.Infof("write to influxdb", measurement, tags, fields)
	point, err := client.NewPoint(measurement, tags, fields, t)
	if err != nil {
		logrus.Errorf("influxdb new point error: %v", err)
		return false, err
	}
	batchpoints.AddPoint(point)
	// Write the batch
	if err := cli.Write(batchpoints); err != nil {
		logrus.Errorf("influxdb write error: %v", err)
		return false, err
	}
	return true, nil
}
