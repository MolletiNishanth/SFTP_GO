package main

import (
	_ "main.go/docs"							
	"net/http"
	 "net"
	 "fmt"
	 "os"
	// "io/ioutil"
     "log"
	"strings"
	"context"
	"cloud.google.com/go/storage"
	"io"
	"golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/agent"
	"github.com/pkg/sftp"
	"github.com/gin-gonic/gin"
	"errors"
	"path/filepath"
	"mime/multipart"
	"time"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

)
// @title SFTP TO GCS API 
// @version 1.0
// @description This is an API which takes input from the api trigger and checks if a text file is present and sends the data to GCS
// @termsOfService demo.com

// @contact.name API Support
// @contact.url http://demo.com/support

// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
const (
	projectID  = "gcp-infra-test-321021"  // FILL IN WITH YOURS
	bucketName = "nishanth_bucket" // FILL IN WITH YOURS
)

type ClientUploader struct {
	cl         *storage.Client
	projectID  string
	bucketName string
	uploadPath string
}

var uploader *ClientUploader
func init() {
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/Users/molleti.nishanth/Downloads/gcp-infra-test-321021-ba44e3282946.json") // FILL IN WITH YOUR FILE PATH
	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	uploader = &ClientUploader{
		cl:         client,
		bucketName: bucketName,
		projectID:  projectID,
		uploadPath: "test-files/",
	}

}
type files struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

func filesById(c *gin.Context) {
	filess := c.Param("id")
	_, err := getfilesById(filess)

	if filess=="sample.txt" {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "File found."})
		// f, err := c.FormFile("file_input")
		// if err != nil {
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"error": err.Error(),
		// 	})
		// 	return
		// }

		blobFile, err := /*f.Open()*/ os.Open("sample.txt")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		err = uploader.UploadFile(blobFile, "sample.txt")			// "./sample.txt" is put in place of f.Filename
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "success",
		})
		return
	}
	if err!=nil {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "file not found !! "})
		return 
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message ": "File not found "})
}

func getfilesById(id string) ( *files, error) {
	
	if _, err := os.Stat("sample.txt"); err == nil {
		if "sample.txt"==id {                           //how to typecast id into file type 
		fmt.Println("File exists !!! ");

		user := "Nishanth, Molleti"
		pass/*, _*/ := "Drowssap@271199"
	
		// Parse Host and Port
		host := "localhost"
		// Default SFTP port
		port := 22
	
	//    hostKey := getHostKey(host)
	
		fmt.Fprintf(os.Stdout, "Connecting to %s ...\n", host)
	
		var auths []ssh.AuthMethod
	
		// Try to use $SSH_AUTH_SOCK which contains the path of the unix file socket that the sshd agent uses 
		// for communication with other processes.
		if aconn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
			auths = append(auths, ssh.PublicKeysCallback(agent.NewClient(aconn).Signers))
		}
	
		// Use password authentication if provided
		if pass != "" {
			auths = append(auths, ssh.Password(pass))
		}
		
		// Initialize client configuration
		config := ssh.ClientConfig{
			User: user,
			Auth: auths,
			// Uncomment to ignore host key check
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		  //  HostKeyCallback: ssh.FixedHostKey(hostKey),
		}
	
		addr := fmt.Sprintf("%s:%d", host, port)
	
		// Connect to server
		conn, err := ssh.Dial("tcp", addr, &config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connecto to [%s]: %v\n", addr, err)
			os.Exit(1)
		}
	
		defer conn.Close()
	
		// Create new SFTP client
		sc, err := sftp.NewClient(conn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to start SFTP subsystem: %v\n", err)
			os.Exit(1)
		}
		defer sc.Close()


		uploadFile(*sc, "./sample.txt", "./remote.txt")
		downloadFile(*sc, "./remote.txt", "./local.txt")
		
		
	 }
	 } else {
		
		fmt.Printf("File does not exist");
	 }

	return nil, errors.New("files not found")
}
func downloadFile(sc sftp.Client, remoteFile, localFile string) (err error) {

    fmt.Fprintf(os.Stdout, "Downloading [%s] to [%s] ...\n", remoteFile, localFile)
    // Note: SFTP To Go doesn't support O_RDWR mode
    srcFile, err := sc.OpenFile(remoteFile, (os.O_RDONLY))
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to open remote file: %v\n", err)
        return
    }
    defer srcFile.Close()

    dstFile, err := os.Create(localFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to open local file: %v\n", err)
        return
    }
    defer dstFile.Close()

    bytes, err := io.Copy(dstFile, srcFile)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Unable to download remote file: %v\n", err)
        os.Exit(1)
    }
    fmt.Fprintf(os.Stdout, "%d bytes copied\n", bytes)
    
    return
}
func uploadFile(sc sftp.Client, localFile, remoteFile string) (err error) {
	fmt.Fprintf(os.Stdout, "Uploading [%s] to [%s] ...\n", localFile, remoteFile)

	srcFile, err := os.Open(localFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open local file: %v\n", err)
		return
	}
	defer srcFile.Close()

	// Make remote directories recursion
	parent := filepath.Dir(remoteFile)
	path := string(filepath.Separator)
	dirs := strings.Split(parent, path)
	for _, dir := range dirs {
		path = filepath.Join(path, dir)
		sc.Mkdir(path)
	}

	// Note: SFTP To Go doesn't support O_RDWR mode
	dstFile, err := sc.OpenFile(remoteFile, (os.O_WRONLY|os.O_CREATE|os.O_TRUNC))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to open remote file: %v\n", err)
		return
	}
	defer dstFile.Close()

	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to upload local file: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d bytes copied\n", bytes)
	
	return
}
func (c *ClientUploader) UploadFile(file multipart.File, object string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := c.cl.Bucket(c.bucketName).Object(c.uploadPath + object).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}

	return nil
}
func main()  {
	// router := gin.Default()
	// router.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"data": "Welcome To Sample program swagger"})
	// })
	// v1 := router.Group("")
	// {
	// 	acc := v1.Group("/functions")
	// 	{
	// 		acc.GET("/pullfile/:id",filesById)
	// 	}
	// }
	//  router.GET("/pullfile/:id",filesById)
	//  router.GET("/swagger/*any",ginSwagger.WrapHandler(swaggerFiles.Handler))
	//  router.Run("localhost:8080")
	r :=setupRouter()
	_ =r.Run("localhost:8080")
	log.Println("router")

}
func setupRouter() *gin.Engine {

	r := gin.New()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "Welcome To Sample program swagger"})
	})

	v1 := r.Group("/")
	{
		accounts := v1.Group("/")
		{
			// accounts.POST("/create", controller.CreateAccount)
			// accounts.PATCH("/update/:id", controller.UpdateAccount)
			// accounts.DELETE("/delete/:id", controller.DeleteAccount)
				fmt.Println("account entered")
				accounts.GET("/pullfile/:id",filesById)
				fmt.Println("account exited")

		}
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r

}



