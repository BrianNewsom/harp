package main

import (
	"fmt"
	"sort"
	"strings"
)

func lsRollbackVersions(servers []*Server, verbose bool) {
	for _, s := range servers {
		fmt.Println("# ====================================")
		fmt.Println("#", s.String())
		s.initPathes()
		releases := s.retrieveAllReleases()
		for _, r := range releases {
			fmt.Println("Release:", r)
			if !verbose {
				continue
			}

			session := s.getSession()
			output, err := session.CombinedOutput(fmt.Sprintf("cat %s/harp/%s/releases/%s/harp-build.info", s.Home, cfg.App.Name, r))
			if err != nil {
				fmt.Printf("echo $HOME on %s error: %s\n", s, err)
			}
			session.Close()
			info := strings.Replace(string(output), "\n", "\n\t", -1)
			fmt.Println("\t" + info[:len(info)-2])
		}
	}
}

func rollback(servers []*Server, version string) {
	for _, s := range servers {
		fmt.Printf("%s rollback start\n", s.String())
		s.initPathes()
		session := s.getSession()
		if option.debug {
			fmt.Println(s.retrieveRollbackScript())
		}
		output, err := session.CombinedOutput(fmt.Sprintf("%s/harp/%s/rollback.sh %s", s.Home, cfg.App.Name, version))
		if err != nil {
			fmt.Printf("rollback on %s error: %s\n", s, err)
		}
		session.Close()
		fmt.Printf(string(output))
		fmt.Printf("%s rollback done\n", s.String())
	}
}

func (s *Server) trimOldReleases() {
	s.initPathes()
	releases := s.retrieveAllReleases()

	if len(releases) <= cfg.RollbackCount {
		return
	}

	for _, release := range releases[:len(releases)-cfg.RollbackCount] {
		session := s.getSession()
		script := fmt.Sprintf("rm -rf %s/harp/%s/releases/%s", s.Home, cfg.App.Name, release)
		if option.debug {
			fmt.Printf("%s: %s\n", s, script)
		}
		output, err := session.CombinedOutput(script)
		if err != nil {
			exitf("failed to exec %s: %s %s", script, output, err)
		}
	}
}

func (s *Server) retrieveAllReleases() []string {
	s.initPathes()
	session := s.getSession()
	rawReleases, err := session.CombinedOutput(fmt.Sprintf(`ls -1 %s/harp/%s/releases`, s.Home, cfg.App.Name))
	if err != nil {
		exitf("failed to exec ls -l: %s %s", rawReleases, err)
	}
	releases := strings.Split(string(rawReleases), "\n")
	var newReleases []string
	for i := range releases {
		if r := strings.TrimSpace(releases[i]); r != "" {
			newReleases = append(newReleases, r)
		}
	}
	releases = newReleases
	sort.Sort(sort.StringSlice(releases))

	return releases
}
