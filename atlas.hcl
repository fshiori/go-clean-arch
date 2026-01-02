// Atlas migration configuration
// Atlas is a modern database schema management tool that supports versioned migrations
// and declarative schema definitions.
//
// Documentation: https://atlasgo.io/

// Define available environments
env "local" {
  // Source: The current state of your database schema (from migration files)
  src = "file://migrations?format=atlas"

  // URL: Database connection string
  // Can be overridden by ATLAS_URL environment variable
  url = env("DATABASE_URL")

  // Dev database for schema validation and diffing
  dev = "docker://postgres/15/dev?search_path=public"

  // Migration directory
  migration {
    dir = "file://migrations"
  }

  // Schema comparison settings
  diff {
    skip {
      drop_schema = true
      drop_table  = false
    }
  }
}

env "dev" {
  src = "file://migrations?format=atlas"

  // For development, use config from environment or defaults
  url = env("DATABASE_URL")

  dev = "docker://postgres/15/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }
}

env "production" {
  src = "file://migrations?format=atlas"

  // Production database URL must be set via environment variable
  url = env("DATABASE_URL")

  migration {
    dir = "file://migrations"
    // Baseline version - set this to your current production version
    baseline = env("ATLAS_BASELINE")
  }

  // Production safety: require approval for destructive changes
  diff {
    skip {
      drop_schema = true
      drop_table  = false
    }
  }

  // Enable auto-approval only if explicitly set
  auto_approve = env("ATLAS_AUTO_APPROVE")
}

// Default environment
env "default" {
  for_each = toset(["postgres", "mysql", "sqlite"])
  url      = each.value == "postgres" ? env("DATABASE_URL") : ""

  src = "file://migrations?format=atlas"

  migration {
    dir = "file://migrations"
  }
}
