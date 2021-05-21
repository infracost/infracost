provider "google" {
  credentials = "{\"type\":\"service_account\"}"
  region      = "us-central1"
}

resource "google_sql_database_instance" "sql_server" {
  name             = "master-instance"
  database_version = "SQLSERVER_2017_ENTERPRISE"
  settings {
    tier              = "db-custom-16-61440"
    availability_type = "ZONAL"
  }
}

resource "google_sql_database_instance" "custom_postgres" {
  name             = "master-instance"
  database_version = "POSTGRES_11"

  settings {
    tier = "db-custom-2-13312"
  }
}

resource "google_sql_database_instance" "HA_custom_postgres" {
  name             = "master-instance"
  database_version = "POSTGRES_11"

  settings {
    tier              = "db-custom-16-61440"
    availability_type = "REGIONAL"
  }
}

resource "google_sql_database_instance" "HA_small_mysql" {
  name             = "master-instance"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-g1-small"
    availability_type = "REGIONAL"
    disk_size         = "100"
  }
}

resource "google_sql_database_instance" "small_mysql" {
  name             = "master-instance"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-g1-small"
    availability_type = "ZONAL"
  }
}

resource "google_sql_database_instance" "micro_mysql_SSD_storage" {
  name             = "master-instance"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
  }
}

resource "google_sql_database_instance" "micro_mysql_HDD_storage" {
  name             = "master-instance"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-f1-micro"
    availability_type = "ZONAL"
    disk_type         = "PD_HDD"
  }
}

resource "google_sql_database_instance" "mysql_standard" {
  name             = "master-instance"
  database_version = "MYSQL_5_7"
  settings {
    tier = "db-n1-standard-32"
  }
}

resource "google_sql_database_instance" "mysql_highmem" {
  name             = "master-instance"
  database_version = "MYSQL_5_7"
  settings {
    tier = "db-n1-highmem-8"
  }
}

resource "google_sql_database_instance" "with_replica" {
  name             = "master-instance"
  database_version = "POSTGRES_11"
  settings {
    tier              = "db-custom-16-61440"
    availability_type = "REGIONAL"
    disk_size         = 500
  }
  replica_configuration {
    username = "replica"
  }
}

resource "google_sql_database_instance" "usage" {
  name             = "master-instance"
  database_version = "MYSQL_8_0"

  settings {
    tier              = "db-g1-small"
    availability_type = "ZONAL"
  }
}
