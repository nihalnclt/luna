package cmd

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/nihalnclt/luna/pkg/http"
	"github.com/spf13/cobra"
)

type Dependency struct {
	Name         string
	Reference    string
	Dependencies []*Dependency
}

// var dependencyMap sync.Map
var dependencyMap = make(map[string]*Dependency)
var depMapMutex sync.Mutex
var resolvedCount, downloadedCount int32

// getPinnedReference retrieves the maximum satisfying version for a given package
func (dep *Dependency) getPinnedReference() error {
	// Check if the reference is a valid, pinned version
	if _, err := semver.NewVersion(dep.Reference); err == nil {
		return nil
	}

	if dep.Reference == "latest" {
		return nil
	}

	parsedRef, err := semver.NewConstraint(dep.Reference)
	if err != nil {
		return fmt.Errorf("invalid version constraint %s for package %s: %w", dep.Reference, dep.Name, err)
	}

	packageData, err := http.PackageData(dep.Name)
	if err != nil {
		return fmt.Errorf("error fetching package data for %s: %w", dep.Name, err)
	}

	if len(packageData.Versions) == 0 {
		return fmt.Errorf("no versions found for package %s", dep.Name)
	}

	var max *semver.Version
	// Find the maximum satisfying version
	for versionStr := range packageData.Versions {
		version, err := semver.NewVersion(versionStr)
		if err != nil {
			continue
		}

		if parsedRef.Check(version) && (max == nil || version.GreaterThan(max)) {
			max = version
		}
	}

	if max == nil {
		return fmt.Errorf("no satisfying version found for package %s", dep.Name)
	}

	dep.Reference = max.String()
	return nil
}

// buildDependencyTree constructs the dependency tree for a given package
func buildPackageDependencyTree(pkg *Dependency, wg *sync.WaitGroup, resolveCh chan<- *Dependency) {
	defer wg.Done()

	if pkg.Name == "ansi-regex" {
		fmt.Println("ansi-regex", pkg.Reference)
	}

	// need to check the pkg.Reference satisfying the dependency
	cacheKey := fmt.Sprintf("%s@%s", pkg.Name, pkg.Reference)
	depMapMutex.Lock()
	if cachedPkg, ok := dependencyMap[cacheKey]; ok {
		depMapMutex.Unlock()
		resolveCh <- cachedPkg
		return
	}

	// TODO:
	// ^1.0.0  -> ^1.0.1 and there is already 1.0.2 in cache, so i can take it from there
	// Pin the reference if it's not already pinned
	if err := pkg.getPinnedReference(); err != nil {
		depMapMutex.Unlock()
		fmt.Println(err.Error())
		return
	}

	versionData, err := http.VersionData(pkg.Name, pkg.Reference)
	if err != nil {
		depMapMutex.Unlock()
		fmt.Printf("error fetching version data for %s: %w", pkg.Name, err)
		return
	}

	newPkg := &Dependency{Name: versionData.Name, Reference: versionData.Version}

	dependencyMap[cacheKey] = newPkg
	depMapMutex.Unlock()

	// Increment the resolved count automatically
	atomic.AddInt32(&resolvedCount, 1)

	fmt.Printf("%sResolving%s %s %s\n", string(Green), string(Reset), newPkg.Name, pkg.Reference)

	var subWg sync.WaitGroup
	subResolveCh := make(chan *Dependency, len(versionData.Dependencies))

	// Resolve and build dependency tree for each sub-dependency
	for depName, depRef := range versionData.Dependencies {
		subDep := &Dependency{Name: depName, Reference: depRef}
		subWg.Add(1)
		go buildPackageDependencyTree(subDep, &subWg, resolveCh)
	}

	go func() {
		subWg.Wait()
		close(subResolveCh)
	}()

	for resolvedDep := range subResolveCh {
		if resolvedDep != nil {
			newPkg.Dependencies = append(newPkg.Dependencies, resolvedDep)
		}
	}

	resolveCh <- newPkg
}

// resolve all packages and it's dependencies
// loop through packages and install them
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a package",
	Long:  "Install a package along with its dependencies",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		start := time.Now()

		packageName := args[0]
		fmt.Println("Resolving Packages...")

		rootPkg := &Dependency{Name: packageName, Reference: "latest"}

		var wg sync.WaitGroup
		resolveCh := make(chan *Dependency, 1)

		wg.Add(1)
		go buildPackageDependencyTree(rootPkg, &wg, resolveCh)

		go func() {
			wg.Wait()
			close(resolveCh)
		}()

		for _ = range resolveCh {
			// fmt.Printf("Resolved Package: %s %s\n", resolvedPkg.Name, resolvedPkg.Reference)
		}

		// if err != nil {
		// 	fmt.Println(err.Error())
		// 	os.Exit(1)
		// }

		// fmt.Println("package name:", depTree.Name)
		// fmt.Println("package version:", depTree.Reference)
		// fmt.Println("package version:", dependencyMap)

		elapsed := time.Since(start).Seconds()
		fmt.Println("Resolved packages", resolvedCount)
		fmt.Printf("✨ Finished in %.2fs.\n", elapsed)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	// installCmd.Flags().StringP("save", "S", "", "Save package to package.json")
	// installCmd.Flags().StringP("save-dev", "D", "", "Save package to package.json as a development dependency")
	// installCmd.Flags().StringP("save-optional", "O", "", "Save package to package.json as an optional dependency")
	// installCmd.Flags().StringP("save-exact", "E", "", "Save package to package.json with an exact version")
	// installCmd.Flags().StringP("save-tilde", "T", "", "Save package to package.json with a tilde version range")
	// installCmd.Flags().StringP("save-peer", "P", "", "Save package to package.json as a peer dependency")
	// installCmd.Flags().StringP("save-workspace",
}
