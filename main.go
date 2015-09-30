package main

import (
    "flag"
    "fmt"
    "io"
    "path/filepath"
    "os"
    "os/exec"

    "github.com/fatih/color"
    "github.com/Varjelus/archivist"
)

// Versions
const (
    VERSION = "0.1.0"
    NW_VERSION = "1.2.0"
    AR_VERSION = "0.9.0"
)

// Defaults
const (
    DEFAULT_SRC = "."
    DEFAULT_REL = "nw-release"
    DEFAULT_ICO = "icon.ico"
)
var DEFAULT_TMP = os.TempDir()
var DEFAULT_TOL = func() string {
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
        fatalError(err.Error())
    }
    return filepath.Join(dir, "buildTools")
}()

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
    nwDir           = flag.String("nw", fmt.Sprintf("%s%cnw", DEFAULT_TOL, os.PathSeparator), "Set NW.js path.")
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
    fatalRed.Println(err)
    os.Exit(2)
}

func copyFile(src, dst string) error {
    // Check first file
    sfi, err := os.Stat(src)
    if err != nil {
        return err
    }
    if !sfi.Mode().IsRegular() {
        // cannot copy non-regular files (e.g., directories,
        // symlinks, devices, etc.)
        return fmt.Errorf("copyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
    }

    // Check second file
    dfi, err := os.Stat(dst)
    if err != nil {
        if !os.IsNotExist(err) {
            return err
        }
    } else {
        if !(dfi.Mode().IsRegular()) {
            return fmt.Errorf("copyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
        }
        if os.SameFile(sfi, dfi) {
            return nil
        }
    }

    // Open source file
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()

    // Create destination file
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()

    // Copy
    if _, err = io.Copy(out, in); err != nil {
        return err
    }

    // Sync
    return out.Sync()
}

// First argument is the final destination path, the rest are files to combine
func copyPlus(dst string, srcs ...string) error {
    temp := filepath.Join(os.TempDir(), "bob.plus")
    together, err := os.OpenFile(temp, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
    if err != nil {
        return err
    }
    defer together.Close()

    for _, src := range srcs {
        f, err := os.Open(src)
        if err != nil {
            return err
        }

        if _, err = io.Copy(together, f); err != nil {
            return err
        }

        f.Close()
    }

    if err := together.Close(); err != nil {
        return err
    }

    return copyFile(temp, dst)
}

func copyWalk(path string, info os.FileInfo, err error) error {
    if err != nil { return err }

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

    dstPath := filepath.Join(*outDir, info.Name())

    // Check if the destination file exists
    dstInfo, err := os.Stat(dstPath)
    // If it does and it errors, return
    if err != nil && os.IsExist(err) {
        return err
    }
    // If it DOES exist and has the same size, return nil
    if os.IsExist(err) && os.SameFile(info, dstInfo) {
        return nil
    }

    return copyFile(path, dstPath)
}

func createZip() {
    abs, err := filepath.Abs(*srcDir)
    if err != nil {
        fatalError(err.Error())
    }
    if err := archivist.Store(abs, filepath.Join(*tmp, "bob.nw")); err != nil {
        fatalError(err.Error())
    }
}

func createExe() {
    if err := os.MkdirAll(*outDir, 0666); err != nil {
        fatalError(err.Error())
    }
    if err := copyFile(filepath.Join(*nwDir, "nw.exe"), exePath); err != nil {
        fatalError(err.Error())
    }
    if err := copyPlus(exePath, exePath, filepath.Join(*tmp, "bob.nw")); err != nil {
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

    resourcer, err := exec.LookPath(fmt.Sprintf("%s/ar/Resourcer.exe", DEFAULT_TOL))
    if err != nil {
        return fmt.Errorf("Can't find the Anolis Resourcer executable: %s", err.Error())
    }
    srcParam := fmt.Sprintf("-src:%s", exePath)
    icoParam := fmt.Sprintf("-file:%s", *icon)
    cmd := exec.Command(resourcer, "-op:upd", srcParam, "-type:14", "-name:IDR_MAINFRAME", icoParam)
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("Can't embed the icon resource: %s", err.Error())
    }

    return nil
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

    // Create the exe
    fmt.Print("Creating the executable... ")
    createExe()
    green.Println("OK")

    // Embed the icon
    fmt.Print("Embedding icon... ")
    if err := createIcon(); err != nil {
        yellow.Println(err.Error())
    } else { green.Println("OK") }

    // Copy files
    fmt.Print("Copying NW.js files... ")
    if err := filepath.Walk(*nwDir, copyWalk); err != nil {
        fatalError("Error copying NW.js files: " + err.Error())
    }
    green.Println("OK")

    // Delete temporary files
    fmt.Print("Deleting temporary files... ")
    if err := os.Remove(filepath.Join(os.TempDir(), "bob.plus")); err != nil {
        fatalError(err.Error())
    }
    if err := os.Remove(filepath.Join(*tmp, "bob.nw")); err != nil {
        fatalError(err.Error())
    }
    green.Println("OK")

    cyan.Println("Done!")
}
