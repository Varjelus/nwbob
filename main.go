package main

import (
    "flag"
    "fmt"
    "io"
    "path/filepath"
    "os"
    "os/exec"
    "strings"

    "github.com/fatih/color"
    "github.com/Varjelus/kopsa"
    "github.com/Varjelus/archivist"
)

// Versions
const (
    VERSION = "0.1.0"
    NW_VERSION = "0.12.3"
    AR_VERSION = "0.9.0"
)

// Defaults
const (
    DEFAULT_SRC = "."
    DEFAULT_REL = "nw-release"
    DEFAULT_ICO = "icon.ico"
)
var (
    DEFAULT_NWF = func() string {
        if _, err := os.Stat("nw"); os.IsNotExist(err) {
            return "nw.zip"
        }
        return "nw" // Use unzipped version if possible
    }()
    DEFAULT_TMP = os.TempDir()
    DEFAULT_TOL = func() string {
        dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
        if err != nil {
            fatalError(err.Error())
        }
        return filepath.Join(dir, "buildTools")
    }()
)

// Colors
var (
    fatalRed    = color.New(color.BgRed)
	red         = color.New(color.FgRed)
	yellow      = color.New(color.FgYellow)
	cyan        = color.New(color.FgCyan)
	green       = color.New(color.FgGreen)
)

// Platforms
var platforms = []string{"win64"}

// Flags
var (
    helpFlag        = flag.Bool("help", false, "Show possible arguments and their default values and explanations.")
    platformsFlag   = flag.Bool("targets", false, "Show available target platforms.")
    versionFlag     = flag.Bool("version", false, "Show the versions and required versions of external tools.")
    nwDir           = flag.String("nw", fmt.Sprintf("%s%c%s", DEFAULT_TOL, os.PathSeparator, DEFAULT_NWF), "Set NW.js path.")
    deflate         = flag.Bool("cmp", false, "Should the ZIP be compressed.")
    srcDir          = flag.String("src", DEFAULT_SRC, "Source directory containing all your project files, including package.json.")
    outDir          = flag.String("out", DEFAULT_REL, "Output directory where you want the packaged application.")
    icon            = flag.String("icon", DEFAULT_ICO, "Desired application icon path.")
    target          = flag.String("target", "win64", "Target platform.")
    projectName     = flag.String("name", "app", "Application name.")
    tmp             = flag.String("tmp", DEFAULT_TMP, "Temporary files directory.")
    usePdf          = flag.Bool("pdf", false, "Does your app need PDF capabilities via pdf.dll? (Large DLL)")
)

var exePath string
var NW_OMIT = []string{"nwjc.exe", "nw.exe", "credits.html"}

func init() {
    flag.Parse()
    if *helpFlag {
		flag.PrintDefaults()
		os.Exit(0)
    }

    if *platformsFlag {
		fmt.Println("Available target platforms:")
        for i := range platforms {
            fmt.Printf("\t* %s", platforms[i])
        }
		os.Exit(0)
    }

    if *versionFlag {
		fmt.Printf("nodebob-go version %s\n", VERSION)
        fmt.Printf("nw.js version %s\n", NW_VERSION)
        fmt.Printf("Anolis Resourcer version %s\n", AR_VERSION)
		os.Exit(0)
    }

    exePath = filepath.Join(*outDir, fmt.Sprintf("%s.exe", *projectName))
}

func fatalError(err string) {
    cleanUp()
    fatalRed.Println(err)
    os.Exit(2)
}

// Copy NW.js files
func copyWalk(path string, info os.FileInfo, err error) error {
    if err != nil { return err }

    fileName := strings.TrimPrefix(path, *nwDir)
    if len(fileName) < 1 {
        return nil
    }
    dstPath := filepath.Join(*outDir, fileName)

    // Check if the destination file exists
    dstInfo, err := os.Stat(dstPath)
    // If it does exist and it errors, return
    if err != nil && os.IsExist(err) {
        return err
    }
    // If it does exist and it's the same file, return nil
    if os.IsExist(err) && os.SameFile(info, dstInfo) {
        return nil
    }

    if info.IsDir() {
        if err := os.Mkdir(dstPath, info.Mode()); err != nil {
            return err
        }
        return nil
    }

    // Keep only regular, nonempty files
    if !info.Mode().IsRegular() || info.Size() == 0 {
        return nil
    }

    // Omit pdf.dll unless otherwise specified
    if info.Name() == "pdf.dll" && !*usePdf {
        return nil
    }

    // Omit specified files
    for _, e := range NW_OMIT {
        if e == info.Name() {
            return nil
        }
    }

    if _, err = kopsa.Copy(dstPath, path); err != nil {
        return err
    }

    return nil
}

func createZip() {
    // Make sure *srcDir/package.json exists
    if _, err := os.Stat(filepath.Join(*srcDir, "package.json")); err != nil {
        if os.IsNotExist(err) {
            fatalError(fmt.Sprintf("%s/package.json not found", *srcDir))
        }
    }

    if err := archivist.Zip(*srcDir, filepath.Join(*tmp, "bob.nw")); err != nil {
        fatalError(err.Error())
    }
}

func createExe() {
    if err := os.RemoveAll(*outDir); err != nil && os.IsExist(err) {
        fatalError(err.Error())
    }

    if err := os.MkdirAll(*outDir, os.ModeDir); err != nil {
        fatalError(err.Error())
    }

    if _, err := kopsa.Copy(exePath, filepath.Join(*nwDir, "nw.exe")); err != nil {
        fatalError(err.Error())
    }

    if err := createIcon(); err != nil {
        yellow.Println(err.Error())
    }

    if _, err := kopsa.Copy(exePath, exePath, filepath.Join(*tmp, "bob.nw")); err != nil {
        fatalError("Can't copy app.nw+nw.exe: " + err.Error())
    }
}

func createIcon() error {
    if _, err := os.Stat(*icon); err != nil {
        if os.IsNotExist(err) {
            return fmt.Errorf("Can't find '%s'", *icon)
        }
        return fmt.Errorf("Can't read '%s': %s", *icon, err.Error())
    }

    resourcer, err := exec.LookPath(filepath.Join(DEFAULT_TOL, "ar", "Resourcer.exe"))
    if err != nil {
        return fmt.Errorf("Can't find the Anolis Resourcer executable: %s", err.Error())
    }
    srcParam := fmt.Sprintf("-src:%s", exePath)
    icoParam := fmt.Sprintf("-file:%s", *icon)
    cmd := exec.Command(resourcer, "-op:upd", srcParam, "-type:14", "-name:IDR_MAINFRAME", icoParam)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Can't embed icon: %s", err.Error())
    }

    return nil
}

func cleanUp() {
    temps := []string{"bobunzip", "bob.nw"}
    for _, temp := range temps {
        if err := os.RemoveAll(filepath.Join(*tmp, temp)); err != nil {
            yellow.Println(err.Error())
        }
    }
}

func main() {
    cyan.Println("nodebob-go\n---\n\n")
    fmt.Printf("NW.js files directory: %s\n", *nwDir)
    fmt.Printf("Source files directory: %s\n", *srcDir)
    fmt.Printf("Output directory: %s\n", *outDir)
    fmt.Printf("Icon path: %s\n", *icon)
    fmt.Printf("Output executable path: %s \n", exePath)
    fmt.Printf("Temporary files directory: %s\n\n", *tmp)

    // Make a ZIP
    fmt.Print("Packaging sources... ")
    createZip()
    green.Println("OK")

    // Unzip NW.js files if necessary
    if filepath.Ext(*nwDir) == ".zip" {
        fmt.Print("Unpacking NW.js files... ")
        if err := archivist.Unzip(*nwDir, filepath.Join(*tmp, "bobunzip")); err != nil {
            fatalError(err.Error())
        }
        *nwDir = filepath.Join(*tmp, "bobunzip")

        green.Println("OK")
    }

    // Create the exe
    fmt.Print("Creating the executable... ")
    createExe()
    green.Println("OK")

    // Copy files
    fmt.Print("Copying NW.js files... ")
    if err := filepath.Walk(*nwDir, copyWalk); err != nil {
        fatalError("Error copying NW.js files: " + err.Error())
    }
    green.Println("OK")

    // Delete temporary files
    fmt.Print("Deleting temporary files... ")
    cleanUp()
    green.Println("OK")

    cyan.Println("Done!")
}
