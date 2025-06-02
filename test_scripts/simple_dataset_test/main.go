package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.viam.com/rdk/app"
	"go.viam.com/rdk/logging"
	"gocv.io/x/gocv"
)

func main() {
	fmt.Println("=== Simple Dataset Test ===")

	// Setup logger
	logger := logging.NewLogger("simple-dataset-test")

	// Dataset configuration - you'll need to provide these values
	apiKey := os.Getenv("VIAM_API_KEY")
	apiKeyID := os.Getenv("VIAM_API_KEY_ID")
	organizationID := os.Getenv("VIAM_ORG_ID")
	datasetID := os.Getenv("VIAM_DATASET_ID")

	if apiKey == "" || apiKeyID == "" || organizationID == "" || datasetID == "" {
		fmt.Println("Dataset test requires environment variables:")
		fmt.Println("  VIAM_API_KEY - Your Viam API key")
		fmt.Println("  VIAM_API_KEY_ID - Your Viam API key ID")
		fmt.Println("  VIAM_ORG_ID - Your Viam organization ID")
		fmt.Println("  VIAM_DATASET_ID - Dataset ID to test with")
		fmt.Println()
		fmt.Println("Get these from: https://app.viam.com")
		fmt.Println("Example:")
		fmt.Println("  export VIAM_API_KEY=\"your-api-key\"")
		fmt.Println("  export VIAM_API_KEY_ID=\"your-api-key-id\"")
		fmt.Println("  export VIAM_ORG_ID=\"your-org-id\"")
		fmt.Println("  export VIAM_DATASET_ID=\"your-dataset-id\"")
		return
	}

	fmt.Printf("Testing with dataset: %s\n", datasetID)
	fmt.Printf("Organization: %s\n", organizationID)

	ctx := context.Background()

	// Create Viam app client
	viamClient, err := app.CreateViamClientWithAPIKey(ctx, app.Options{}, apiKey, apiKeyID, logger)
	if err != nil {
		fmt.Printf("Failed to create Viam client: %v\n", err)
		return
	}
	defer viamClient.Close()

	dataClient := viamClient.DataClient()

	// Debug: First try to get ANY images from this organization (without dataset filter)
	fmt.Println("Debug: Checking if we can access any images in this organization...")
	debugFilter := &app.Filter{
		OrganizationIDs: []string{organizationID},
	}

	debugResp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
		Filter: debugFilter,
		Limit:  5,
	})
	if err != nil {
		fmt.Printf("❌ Cannot access organization data: %v\n", err)
		fmt.Println("This suggests an API key permission issue.")
		return
	}
	fmt.Printf("✅ Found %d total images in organization\n", len(debugResp.BinaryData))

	// NEW: Try to fetch images without dataset filter to see what's available
	fmt.Println("Debug: Trying to fetch actual IMAGES from organization (no dataset filter)...")
	orgImageFilter := &app.Filter{
		OrganizationIDs: []string{organizationID},
	}

	orgImageResp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
		Filter: orgImageFilter,
		Limit:  10,
	})
	if err != nil {
		fmt.Printf("❌ Cannot fetch images from organization: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d total data entries in organization\n", len(orgImageResp.BinaryData))

		if len(orgImageResp.BinaryData) > 0 {
			fmt.Println("Data types found in organization:")
			orgDataTypes := make(map[string]int)

			for _, data := range orgImageResp.BinaryData {
				// Check file types
				if data.Metadata != nil {
					var contentType string
					if data.Metadata.FileName != "" {
						if len(data.Metadata.FileName) > 4 {
							ext := data.Metadata.FileName[len(data.Metadata.FileName)-4:]
							switch ext {
							case ".jpg", "jpeg":
								contentType = "image/jpeg"
							case ".png":
								contentType = "image/png"
							case ".bmp":
								contentType = "image/bmp"
							default:
								contentType = "unknown (" + ext + ")"
							}
						} else {
							contentType = "unknown filename"
						}
					} else {
						contentType = "unknown"
					}
					orgDataTypes[contentType]++
				}
			}

			for contentType, count := range orgDataTypes {
				fmt.Printf("  - %s: %d entries\n", contentType, count)
			}

			fmt.Println("This shows there IS data available in your organization.")
			if len(orgDataTypes) > 0 {
				fmt.Println("The issue is specifically with the dataset being empty, not with API access.")
			}
		}
	}

	// Debug: Check if we can list datasets in this organization
	fmt.Println("Debug: Checking available datasets in organization...")
	datasets, err := dataClient.ListDatasetsByOrganizationID(ctx, organizationID)
	if err != nil {
		fmt.Printf("❌ Cannot list datasets: %v\n", err)
	} else {
		fmt.Printf("✅ Found %d datasets in organization:\n", len(datasets))
		for i, dataset := range datasets {
			fmt.Printf("  %d. ID: %s, Name: %s\n", i+1, dataset.ID, dataset.Name)
			if dataset.ID == datasetID {
				fmt.Printf("     👆 This is the target dataset!\n")
			}
		}
	}

	// Create filter for dataset
	filter := &app.Filter{
		DatasetID: datasetID,
	}

	fmt.Println("Now checking specifically for dataset images...")

	// Enhanced debugging: Check what types of data exist in this dataset
	fmt.Println("Debug: Checking ALL data types in this dataset...")

	// Try to get ANY data from this dataset (not just images)
	allDataResp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		fmt.Printf("❌ Failed to fetch any data from dataset: %v\n", err)
		return
	}

	fmt.Printf("Found %d total data entries in dataset\n", len(allDataResp.BinaryData))

	if len(allDataResp.BinaryData) > 0 {
		fmt.Println("Data types found in dataset:")
		dataTypes := make(map[string]int)
		for _, data := range allDataResp.BinaryData {
			if data.Metadata != nil {
				// Check different possible fields for content type
				var contentType string
				if data.Metadata.FileName != "" {
					// Infer from filename extension
					if len(data.Metadata.FileName) > 4 {
						ext := data.Metadata.FileName[len(data.Metadata.FileName)-4:]
						switch ext {
						case ".jpg", "jpeg":
							contentType = "image/jpeg"
						case ".png":
							contentType = "image/png"
						case ".bmp":
							contentType = "image/bmp"
						default:
							contentType = "unknown (" + ext + ")"
						}
					} else {
						contentType = "unknown filename"
					}
				} else {
					contentType = "unknown"
				}
				dataTypes[contentType]++
			}
		}
		for contentType, count := range dataTypes {
			fmt.Printf("  - %s: %d entries\n", contentType, count)
		}
	}

	// Also try specifically filtering for common image data
	fmt.Println("Debug: Checking for data that might be images...")
	// Note: We'll check all data and try to identify images by other means
	// since MIME type filtering might not be available

	fmt.Println("First, getting list of available images...")

	// First, get metadata only (without binary data) to see what's available
	metadataResp, err := dataClient.BinaryDataByFilter(ctx, false, &app.DataByFilterOptions{
		Filter: filter,
		Limit:  10, // Get metadata for up to 10 images
	})
	if err != nil {
		fmt.Printf("Failed to fetch dataset metadata: %v\n", err)
		return
	}

	fmt.Printf("Found %d images in dataset\n", len(metadataResp.BinaryData))

	if len(metadataResp.BinaryData) == 0 {
		fmt.Println("No images found in dataset.")
		fmt.Println("POSSIBLE REASONS:")
		fmt.Println("1. Dataset is empty (contains no data at all)")
		fmt.Println("2. Dataset contains non-image data (sensor readings, logs, etc.)")
		fmt.Println("3. Images are stored with different MIME types than expected")
		fmt.Println("4. Dataset filter is not working as expected")
		fmt.Println()
		fmt.Println("Check the Viam dashboard to verify what data this dataset actually contains.")
		return
	}

	// Now fetch binary data for up to 5 images, one at a time
	fmt.Println("Fetching binary data for images (limit: 5)...")
	maxImages := 5
	if len(metadataResp.BinaryData) < maxImages {
		maxImages = len(metadataResp.BinaryData)
	}

	for i := 0; i < maxImages; i++ {
		metadata := metadataResp.BinaryData[i]
		fmt.Printf("Fetching image %d (ID: %s)...\n", i+1, metadata.Metadata.ID)

		// Fetch binary data for this specific image using BinaryDataByIDs with string slice
		binaryResp, err := dataClient.BinaryDataByIDs(ctx, []string{metadata.Metadata.ID})
		if err != nil {
			fmt.Printf("Failed to fetch binary data for image %d: %v\n", i+1, err)
			continue
		}

		if len(binaryResp) == 0 || binaryResp[0].Binary == nil {
			fmt.Printf("No binary data returned for image %d\n", i+1)
			continue
		}

		binaryData := binaryResp[0]
		fmt.Printf("Processing image %d: %d bytes\n", i+1, len(binaryData.Binary))

		// Try to decode the image
		newFrame, err := gocv.IMDecode(binaryData.Binary, gocv.IMReadColor)
		if err != nil || newFrame.Empty() {
			fmt.Printf("Failed to decode image %d: %v\n", i+1, err)

			// Save raw binary data for inspection
			filename := fmt.Sprintf("dataset_raw_%d.bin", i+1)
			if err := os.WriteFile(filename, binaryData.Binary, 0644); err != nil {
				fmt.Printf("Failed to save raw data: %v\n", err)
			} else {
				fmt.Printf("Saved raw binary data to: %s\n", filename)
			}
			continue
		}

		fmt.Printf("Successfully decoded image %d: %dx%d pixels\n", i+1, newFrame.Cols(), newFrame.Rows())

		// Re-encode as JPEG and save
		buf, err := gocv.IMEncode(".jpg", newFrame)
		if err != nil {
			fmt.Printf("Failed to encode image %d: %v\n", i+1, err)
		} else {
			filename := fmt.Sprintf("dataset_frame_%d.jpg", i+1)
			if err := os.WriteFile(filename, buf.GetBytes(), 0644); err != nil {
				fmt.Printf("Failed to save image %d: %v\n", i+1, err)
			} else {
				fmt.Printf("Saved image to: %s\n", filename)
			}
			buf.Close()
		}

		newFrame.Close()

		// Show metadata if available
		if binaryData.Metadata != nil {
			fmt.Printf("  Timestamp: %v\n", binaryData.Metadata.TimeRequested)
			if binaryData.Metadata.FileName != "" {
				fmt.Printf("  Filename: %s\n", binaryData.Metadata.FileName)
			}
		}

		time.Sleep(500 * time.Millisecond) // Shorter delay between individual requests
	}

	fmt.Println("\n=== Simple Dataset Test Complete ===")
	fmt.Printf("Successfully processed %d images from your dataset.\n", maxImages)
	fmt.Println("This test verifies that you can fetch images from your Viam dataset.")
	fmt.Println("If this works, the dataset mode should work too.")
}
