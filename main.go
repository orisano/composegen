package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/orisano/subflag"
	"github.com/xo/dburl"
)

var defaultPorts = map[string]int{
	"postgres": 5432,
	"mysql":    3306,
	"redis":    6379,
}

type DBCommand struct {
	URL     string
	Tag     string
	Service string
}

func (c *DBCommand) FlagSet() *flag.FlagSet {
	fs := flag.NewFlagSet("db", flag.ExitOnError)
	fs.StringVar(&c.URL, "url", "", "url syntax connection string (required)")
	fs.StringVar(&c.Tag, "tag", "latest", "image tag")
	fs.StringVar(&c.Service, "s", "", "docker-compose service name")
	return fs
}

type service struct {
	Image       string            `yaml:"image"`
	Command     string            `yaml:"command,omitempty"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Ports       []string          `yaml:"ports"`
	CapAdd      []string          `yaml:"cap_add"`
	comments    []string
}

func (c *DBCommand) Run(_ []string) error {
	if c.URL == "" {
		return flag.ErrHelp
	}
	u, err := url.ParseRequestURI(c.URL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	dialect := u.Scheme
	if u, err := dburl.Parse(c.URL); err == nil {
		dialect = u.Unaliased
	}
	defaultPort, ok := defaultPorts[dialect]
	if !ok {
		return fmt.Errorf("unsupported dialect: %s", dialect)
	}

	fmt.Println(`version: '3'`)
	fmt.Println(`services:`)

	port := defaultPort
	if s := u.Port(); s != "" {
		p, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("parse port(%v): %w", s, err)
		}
		port = p
	}

	serviceName := c.Service
	if serviceName == "" && u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1" {
		serviceName = u.Hostname()
	}
	if serviceName == "" {
		serviceName = u.Scheme
	}

	var s service
	s.Ports = []string{
		fmt.Sprintf("%d:%d", port, defaultPort),
	}

	database := strings.TrimPrefix(u.Path, "/")
	username := u.User.Username()
	password, _ := u.User.Password()
	switch dialect {
	case "mysql":
		s.Image = "mysql:" + c.Tag
		s.Environment = map[string]string{
			"MYSQL_DATABASE":             database,
			"MYSQL_USER":                 username,
			"MYSQL_PASSWORD":             password,
			"MYSQL_ALLOW_EMPTY_PASSWORD": "yes",
		}
		s.Command = "--default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --long-query-time=0 --slow-query-log=ON --slow-query-log-file=slow.log"
		if !strings.HasPrefix(c.Tag, "5") {
			s.CapAdd = append(s.CapAdd, "SYS_NICE")
		}
		s.comments = append(s.comments, "  volumes:")
		s.comments = append(s.comments, "  - ./sql:/docker-entrypoint-initdb.d:ro")
	case "postgres":
		s.Image = "postgres:" + c.Tag
		s.Environment = map[string]string{
			"POSTGRES_DB":       database,
			"POSTGRES_USER":     username,
			"POSTGRES_PASSWORD": password,
		}
		s.comments = append(s.comments, "  volumes:")
		s.comments = append(s.comments, "  - ./sql:/docker-entrypoint-initdb.d:ro")
	case "redis":
		s.Image = "redis:" + c.Tag
		if password != "" {
			s.Command = fmt.Sprintf("--requirepass %q", password)
		}
	}
	var buf bytes.Buffer
	err = yaml.NewEncoder(&buf).Encode(map[string]interface{}{serviceName: s})
	if err != nil {
		return fmt.Errorf("encode service: %w", err)
	}
	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		fmt.Printf("%*s%s\n", 2, "", scanner.Text())
	}
	for _, comment := range s.comments {
		fmt.Printf("#%*s%s\n", 1, "", comment)
	}

	return nil
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("composegen: ")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return subflag.SubCommand(os.Args[1:], []subflag.Command{
		&DBCommand{},
	})
}
