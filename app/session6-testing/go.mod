module mini-asm

go 1.23.0

toolchain go1.24.12

require (
	github.com/google/uuid v1.6.0 // UUID generation
	github.com/lib/pq v1.10.9 // PostgreSQL driver
	github.com/spf13/viper v1.21.0
)

require (
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.11.0 // indirect
	github.com/sourcegraph/conc v0.3.1-0.20240121214520-5f936abd7ae8 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.29.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

// 🎓 TEACHING NOTES:
//
// Session 4 uses the same dependencies as Session 3:
//
// 1. github.com/google/uuid
//    - Generate unique IDs for assets
//
// 2. github.com/lib/pq
//    - PostgreSQL driver
//    - database/sql compatible
//
// No new  dependencies needed!
// All advanced features implemented with standard library:
// - Query parameter parsing: net/url
// - Validation: regexp, net packages
// - JSON: encoding/json
//
// This shows Go's powerful standard library!
//
// Installation:
//   go mod tidy
