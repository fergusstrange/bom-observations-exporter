package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jlaffaye/ftp"
	"github.com/mholt/archiver/v3"
	"github.com/olivere/elastic/v7"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const timeFormat = "20060102150405 -07:00"
const bomObservationsIndex = "bom-observations"

var observationFiles = [...]string{"IDD60910.tgz", "IDQ60910.tgz", "IDT60910.tgz", "IDW60910.tgz", "IDN60910.tgz", "IDS60910.tgz", "IDV60910.tgz"}

func Handler(ctx context.Context) error {
	log.Printf("+%v", ctx)

	config, err := LoadConfig()
	if err != nil {
		return err
	}

	tempDirectory, err := ioutil.TempDir("", "bom_observations")
	if err != nil {
		return err
	}

	extractedObservationsDirectory := filepath.Join(tempDirectory, "extracted")

	defer func() {
		if err := os.RemoveAll(tempDirectory); err != nil {
			log.Fatal(err)
		}
	}()

	ftpConnection, err := ftp.Dial("ftp.bom.gov.au:21", ftp.DialWithTimeout(3*time.Second))
	if err != nil {
		return err
	}

	defer func() {
		if err := ftpConnection.Quit(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := ftpConnection.Login("anonymous", "anonymous"); err != nil {
		return err
	}

	for _, observationFile := range observationFiles {
		observationFilePath := "/anon/gen/fwo/" + observationFile
		entries, err := ftpConnection.List(observationFilePath)
		if err != nil {
			return err
		}

		if len(entries) < 1 {
			log.Printf("skipping %s as not found", observationFilePath)
			continue
		}

		response, err := ftpConnection.Retr(observationFilePath)
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadAll(response)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(filepath.Join(tempDirectory, observationFile), bytes, 0660)

		if err := response.Close(); err != nil {
			return err
		}
	}

	downloadedFiles, err := ioutil.ReadDir(tempDirectory)
	if err != nil {
		return err
	}

	for _, f := range downloadedFiles {
		downloadedFile := filepath.Join(tempDirectory, f.Name())
		if filepath.Ext(downloadedFile) != ".tgz" {
			continue
		}

		if err := archiver.Walk(downloadedFile, func(archivedFileName archiver.File) error {
			if filepath.Ext(archivedFileName.Name()) == ".json" {
				if err := archiver.Extract(downloadedFile, archivedFileName.Name(), extractedObservationsDirectory); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	extractedObservationFiles, err := ioutil.ReadDir(extractedObservationsDirectory)
	if err != nil {
		return err
	}

	bomObservations := make(map[int][]ElasticSearchBomObservation, 0)
	for _, extractedObservationFile := range extractedObservationFiles {
		if filepath.Ext(extractedObservationFile.Name()) == ".json" {
			bytes, err := ioutil.ReadFile(filepath.Join(extractedObservationsDirectory, extractedObservationFile.Name()))
			if err != nil {
				return err
			}

			var bomObservationWrapper BomObservationWrapper
			if err := json.Unmarshal(bytes, &bomObservationWrapper); err != nil {
				return err
			}

			for _, bomObservation := range bomObservationWrapper.Observations.Data {
				elasticSearchBomObservation, err := bomObservation.ToElasticSearchBomObservation()
				if err != nil {
					return err
				}
				_, ok := bomObservations[elasticSearchBomObservation.WMO]
				if !ok {
					bomObservations[elasticSearchBomObservation.WMO] = make([]ElasticSearchBomObservation, 0)
				}
				bomObservations[elasticSearchBomObservation.WMO] = append(bomObservations[elasticSearchBomObservation.WMO], elasticSearchBomObservation)
			}
		}
	}

	client, err := newElasticClient(config.ElasticSearchURL)
	if err != nil {
		return err
	}

	for wmo, wmoObservations := range bomObservations {
		observationsToStore, err := filterForObservationsNewerThanMostRecentStored(client, wmo, wmoObservations)
		if err != nil {
			return err
		}

		bulkIndexRequests := make([]elastic.BulkableRequest, 0)
		for _, observation := range observationsToStore {
			bulkIndexRequests = append(bulkIndexRequests, elastic.NewBulkIndexRequest().Doc(observation))
		}

		if len(bulkIndexRequests) > 0 {
			_, err := client.Bulk().Index(bomObservationsIndex).
				Add(bulkIndexRequests...).
				Do(context.Background())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func filterForObservationsNewerThanMostRecentStored(client *elastic.Client, wmo int, wmoObservations []ElasticSearchBomObservation) ([]ElasticSearchBomObservation, error) {
	existingWmoObservationsSearch, err := client.Search(bomObservationsIndex).
		Query(elastic.NewTermQuery("wmo", wmo)).
		Sort("timestamp", false).
		Size(1).
		Do(context.Background())
	if err != nil {
		return wmoObservations, err
	}

	if existingWmoObservationsSearch.TotalHits() > 0 {
		filtered := wmoObservations[:0]
		latestObservation := new(ElasticSearchBomObservation)
		if err := json.Unmarshal(existingWmoObservationsSearch.Hits.Hits[0].Source, latestObservation); err != nil {
			return filtered, err
		}

		for _, observation := range wmoObservations {
			if observation.Timestamp.After(latestObservation.Timestamp) {
				filtered = append(filtered, observation)
			}
		}

		return filtered, nil
	}

	return wmoObservations, nil
}

func main() {
	lambda.Start(Handler)
}
