package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type PackingList struct {
	XMLName        xml.Name  `xml:"PackingList"`
	XMLNS          string    `xml:"xmlns,attr"`
	Id             string    `xml:"Id"`
	AnnotationText string    `xml:"AnnotationText"`
	IssueDate      string    `xml:"IssueDate"`
	Issuer         string    `xml:"Issuer"`
	Creator        string    `xml:"Creator"`
	AssetList      AssetList `xml:"AssetList"`
}

type AssetList struct {
	XMLName xml.Name `xml:"AssetList"`
	Assets  []Asset  `xml:"Asset"`
}

type Asset struct {
	XMLName        xml.Name `xml:"Asset"`
	Id             string   `xml:"Id"`
	AnnotationText string   `xml:"AnnotationText"`
	Hash           string   `xml:"Hash"`
	Size           string   `xml:"Size"`
	Type           string   `xml:"Type"`
}

// To create the assetmap strucure with the necessary variables.
type AssetMap struct {
	XMLName xml.Name `xml:"AssetMap"`   //This will create the root element for the assetmap
	XMLNS   string   `xml:"xmlns,attr"` //This will add the namespace for the assetmap
	//all the this that are provided below is sub elements of the root element
	Id             string      `xml:"Id"`
	AnnotationText string      `xml:"AnnotationText"`
	Creator        string      `xml:"Creator"`
	VolumeCount    string      `xml:"VolumeCount"`
	IssueDate      string      `xml:"IssueDate"`
	Issuer         string      `xml:"Issuer"`
	AssetList      AMAssetList `xml:"AssetList"`
}

// Define the assetmap assetlist structure
type AMAssetList struct {
	XMLName xml.Name  `xml:"AssetList"` //this is the elementname
	Assets  []AMAsset `xml:"Asset"`     //list of assets
}

// Define the asset stucture
type AMAsset struct {
	XMLName        xml.Name  `xml:"Asset"`
	Id             string    `xml:"Id"`
	AnnotationText string    `xml:"AnnotationText"`
	ChunkList      ChunkList `xml:"ChunkList"`
}

// Define the Chunklist structure
type ChunkList struct {
	XMLName xml.Name `xml:"ChunkList"`
	Chunk   []Chunk  `xml:"Chunk"`
}

// Define the chunk structure
type Chunk struct {
	XMLName xml.Name `xml:"Chunk"`
	Path    string   `xml:"Path"`
}

// Map is created to store the uuid for each file in a dictionary.
var myMap = make(map[string]string)

// It iterate over all the files and then generate the uuid for each files and then the value is stored in a map
func iterate(userFolderName string) {
	filepath.Walk(userFolderName, func(userFolderName string, file os.FileInfo, err error) error {
		if err != nil {
			log.Fatalf(err.Error())
		} else {
			myMap[string(file.Name())] = uuid.New().String() // stores it in the myMap
		}
		return nil
	})
}

// It will generate the hash values for each file using SHA1 algorithm.
func findHashValue(file_path string) string {
	file, err := os.Open(file_path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		fmt.Println(err)
	}

	return hex.EncodeToString(hash.Sum(nil)) // in hex format and also encoded as string
}

// The below is the PackingList function. It will create the assetmap file with the provided folderpath and uuid .
func createPackingList(folderPath string, uuidCommon map[string]string) {
	var assets []Asset // Create an empty slice of Asset
	// It will iterate over all files in the folder and create an Asset for each file in a directory.
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			fileName := filepath.Base(path)
			asset := Asset{
				Id:             "urn:uuid:" + uuidCommon[fileName],
				AnnotationText: filepath.Base(folderPath),
				Hash:           findHashValue(path),
				Size:           strconv.FormatInt(info.Size(), 10),
				Type:           mime.TypeByExtension(filepath.Ext(path)),
			}
			assets = append(assets, asset)
		}
		return nil
	})
	//This creates an Packinglis structure
	Packinglist := PackingList{
		XMLNS:          "http://www.smpte-ra.org/schemas/429-8/2007/PKL",
		Id:             "urn:uuid:" + uuidCommon["UUID"],
		AnnotationText: filepath.Base(folderPath),
		IssueDate:      time.Now().UTC().Format("2006-01-02T15:04:05-07:00"),
		Issuer:         "Qube Cinema",
		Creator:        "Qube",
		AssetList: AssetList{
			Assets: assets,
		},
	}

	output, err := xml.MarshalIndent(Packinglist, "", "    ")
	if err != nil {
		panic(err)
	}

	pkfile, err := os.Create(filepath.Join(folderPath, "Packinglist.xml"))
	if err != nil {
		panic(err)
	}
	defer pkfile.Close()

	pkfile.Write([]byte(xml.Header))
	pkfile.Write(output)

}

// The below is the assetmap function. It will create the assetmap file with the provided folderpath and uuid .
func createAssetmap(folderPath string, uuidCommon map[string]string) {
	var assets []AMAsset // Create an empty slice of Asset
	// It will iterate over all files in the folder and create an Asset for each file in a directory.
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		//This code block checks if the current file in the directory is a regular file.It will crate the asset for the regular files only.
		if info.Mode().IsRegular() {
			fileName := filepath.Base(path)
			asset := AMAsset{
				Id:             "urn:uuid:" + uuidCommon[fileName],
				AnnotationText: filepath.Base(folderPath),
				ChunkList: ChunkList{
					Chunk: []Chunk{
						{Path: info.Name()},
					},
				},
			}
			assets = append(assets, asset)
		}
		return nil
	})
	//This creates an AssetMap structure
	AssetMap := AssetMap{
		XMLNS:          "http://www.smpte-ra.org/schemas/429-9/2007/AM",
		Id:             "urn:uuid:" + uuidCommon["UUID"],
		AnnotationText: filepath.Base(folderPath),
		Creator:        "Qube",
		VolumeCount:    "1",
		IssueDate:      time.Now().UTC().Format("2006-01-02T15:04:05-07:00"),
		Issuer:         "Qube Cinema",
		AssetList: AMAssetList{
			Assets: assets,
		},
	}
	//This encodes the AssetMap structure into XML format
	output, err := xml.MarshalIndent(AssetMap, "", "    ")
	if err != nil {
		panic(err)
	}
	//The assetmap file is created to store the file in xml file
	pkfile, err := os.Create(filepath.Join(folderPath, "assetmap.xml"))
	if err != nil {
		panic(err) //if the file is not able to create then it will throws an error
	}
	defer pkfile.Close()

	pkfile.Write([]byte(xml.Header))
	pkfile.Write(output)
}

func main() {
	// To get the folder path from the user.
	var userFolderName string
	fmt.Println("Enter Your Folder path: ")

	// Taking input from user
	fmt.Scanln(&userFolderName)
	// To check the folder path is valid or not.
	_, err := os.Stat(userFolderName)
	if err != nil {
		println("os.Stat(): error for folder name ", userFolderName)
		println("and error is : ", err.Error())
		// The if condition checks the directory exists or not.
		if os.IsNotExist(err) {
			println("Directory Does not exists.")
		}
	} else {
		f, _ := os.Open(userFolderName)
		defer f.Close()

		files, _ := f.Readdirnames(1) // Or f.Readdir(1)
		// To check the directory is empty or not.
		if len(files) == 0 {
			println("your directory is empty")
		} else {
			//It will generate the uuid for each file and store it in a map based on key value pair
			uuidFile := uuid.New().String()
			myMap["UUID"] = uuidFile
			iterate(userFolderName) //iterate function is called.
			// packinglist and the assetmap function is called
			createPackingList(userFolderName, myMap)
			createAssetmap(userFolderName, myMap)
		}

	}
}
