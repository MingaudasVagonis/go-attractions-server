package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/nfnt/resize"
	"github.com/oliamb/cutter"
)

// Function takes in a command split by spaces, merges cache with an external database,
// downloads and processes images and either saves them locally or posts them to the provided
// url in json format.
func merge(parts []string) string {

	attractions, err := readCache()
	if err != nil {
		return fmt.Sprintf("Failed to merge: %s", err.Error())
	}

	// Committing attractions to an external DB
	if err := commitAttractionsToDB(parts[1], attractions); err != nil {
		return fmt.Sprintf("Failed to merge: %s", err.Error())
	}

	var (
		// Slice that contains images yet to download.
		toDownload []Downloadable
		// Slice that contains ids of attractions that either failed or doens't have an url.
		failed []string
	)

	// Extracting ids and urls from attractions.
	getUrls(attractions, &toDownload, &failed)

	// Downloading images.
	download(&toDownload, &failed)

	// Resizing & cropping images to fit required dimensions.
	process(&toDownload, &failed)

	// If provided, images will be send to an url.
	if len(parts) > 2 {
		send(toDownload, parts[2])
		return fmt.Sprintf("Sent %d images to %s.\nFinished with %d failed images\n\t%v",
			len(toDownload), parts[2], len(failed), failed)
	}

	// If no url provided images will be saved locally.
	save(toDownload, &failed)
	return fmt.Sprintf("Finished with %d failed images\n\t%v", len(failed), failed)
}

// Function takes in a reference to a slice of Downloadables and
// a reference to slice of ids that failed to download. Bytes are downloaded
// from an image url and put into the downloadable.
func download(toDownload *[]Downloadable, failed *[]string) {
	// Creating a new slice with the same underlying slice in order to
	// leave out downloadables that failed to download.
	new_down := (*toDownload)[:0]
	for _, down := range *toDownload {

		// Downloading image bytes.
		response, err := http.Get(down.url)

		if err != nil {
			*failed = append(*failed, down.id)
			continue
		} else {
			// Assigning bytes to the downloadable.
			down.image, err = ioutil.ReadAll(response.Body)
			new_down = append(new_down, down)
		}
		response.Body.Close()
	}
	*toDownload = new_down
}

// Function takes in a reference to a slice of Downloadables and
// a reference to slice of ids that failed to download. new image.Image
// object is processed and assigned to the downloadable.
func process(toDownload *[]Downloadable, failed *[]string) {
	// Creating a new slice with the same underlying slice in order to
	// leave out downloadables that failed to process.
	new_down := (*toDownload)[:0]

	for _, down := range *toDownload {

		// Creating image.Image from bytes.
		img, _, err := image.Decode(bytes.NewReader(down.image))

		if err != nil {
			*failed = append(*failed, down.id)
			continue
		}

		// Resizing the image to be 1200 pixels width.
		img = resize.Resize(1200, 0, img, resize.Bicubic)
		// Cropping the image to fit  3 by 2 aspect ration.
		img, err = cutter.Crop(img, cutter.Config{Width: 3, Height: 2, Mode: cutter.Centered, Options: cutter.Ratio})

		if err != nil {
			*failed = append(*failed, down.id)
			continue
		}
		// Assigning image.Image to the downloadable.
		down.decoded_img = img
		new_down = append(new_down, down)
	}
	*toDownload = new_down
}

// Function takes in a slice of attractions, a reference to a slice of Downloadables and
// a reference to slice of ids that failed to download. Attraction's id and url is added to a
// toDownload slice is url is present, otherwise id is added to the failed slice.
func getUrls(attractions []Attraction, toDownload *[]Downloadable, failed *[]string) {
	for _, attr := range attractions {
		if attr.url.Valid {
			*toDownload = append(*toDownload, Downloadable{url: attr.url.String, id: attr.id})
		} else {
			*failed = append(*failed, attr.id)
		}
	}
}

// Function takes in a reference to a slice of Downloadables and
// a reference to slice of ids that failed to download. The image is saved locally as
// a jpeg with compression 80 or added as a failed id.
func save(downloadables []Downloadable, failed *[]string) {
	options := jpeg.Options{Quality: 80}
	for _, down := range downloadables {

		// Creating the file.
		file, err := os.Create(fmt.Sprintf("%s.jpg", down.id))
		if err != nil {
			*failed = append(*failed, down.id)
			continue
		}

		defer file.Close()
		// Encoding the image to the file created.
		if err := jpeg.Encode(file, down.decoded_img, &options); err != nil {
			*failed = append(*failed, down.id)
		}
	}
}

// Function takes in a slice of Downloadables and
// a url string to send the images to. Images are encoded as a
// base64 string.
func send(downloadables []Downloadable, url string) {
	json_data := make([]map[string]string, 0)

	for _, down := range downloadables {
		json_data = append(json_data, map[string]string{
			"id":    down.id,
			"image": base64.StdEncoding.EncodeToString(down.image),
		})
	}

	// Marshalling into json array of json objects {id string, image string}.
	json, _ := json.Marshal(json_data)

	// Posting the json to the url.
	if _, err := http.Post(url, "application/json", bytes.NewBuffer(json)); err != nil {
		fmt.Println("Failed to send images: ", err.Error())
	}

}

type Downloadable struct {
	url         string
	id          string
	image       []byte
	decoded_img image.Image
}
