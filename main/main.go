package main

import (
	
	"net/http"
	 "net"
	 "fmt"
	 "os"
	// "io/ioutil"
    // "log"
	"strings"
	"io"
	"golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/agent"
	"github.com/pkg/sftp"
	"github.com/gin-gonic/gin"
	"errors"
	"path/filepath"
)
type book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	Quantity int    `json:"quantity"`
}

var books = []book{
	{ID: "1", Title: "In Search of Lost Time", Author: "Marcel Proust", Quantity: 2},
	{ID: "2", Title: "The Great Gatsby", Author: "F. Scott Fitzgerald", Quantity: 5},
	{ID: "3", Title: "War and Peace", Author: "Leo Tolstoy", Quantity: 6},
}
// type file struct{
// 	ID          string  `json:"id"`
// 	FILENAME    string  `json:"filename"`
// 	RESPONSE    int     `json:"response"`
// 	DATA        string	`json:"data"`
// }
// var files = []file{
// 	{ID: "1",FILENAME: "file", RESPONSE: 1, DATA: "Hello1"} ,
// 	{ID: "2",FILENAME: "file2", RESPONSE: 2, DATA: "Hello2"} ,
// }
//  func getFiles(c *gin.Context){
//  	c.IndentedJSON(http.StatusOK, files)
//  }
//  func fileData(c *gin.Context){
// 	id:=c.Param("id")
// 	fmt.Println(id)
// 	file,err:=getFileData(id)
// 	 if err != nil{
// 		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Book not found."})
// 		return 
// 	 }
// 	 c.IndentedJSON(http.StatusOK,file)
//  }
//  func getFileData(id string) (*file ,  error){
// 	for i, b := range files{
// 		if b.ID==id {
// 			return &files[i], nil 
// 		}
// 	}
// 	return nil , errors.New("File not found")
// }
func bookById(c *gin.Context) {
	id := c.Param("id")
	book, err := getBookById(id)

	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Book not found."})
		return
	}
	c.IndentedJSON(http.StatusOK, book)
}

func getBookById(id string) ( *book, error) {
	// for i, b := range books {
	// 	if b.ID == id {
	// 		return &books[i], nil
	// 	}
	// }
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

	return nil, errors.New("book not found")
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
func main()  {
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
	router := gin.Default()
	// router.GET("/files",getFiles)
	// router.GET("/files/:id ",fileData)
	 router.GET("/books/:id",bookById)
	 router.Run("localhost:8080")

}
