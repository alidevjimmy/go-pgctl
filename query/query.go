package query

import "fmt"

func CreatePublicationQuery(name string) string {
	return fmt.Sprintf("CREATE PUBLICATION %s FOR ALL TABLES", name)
}

func CreateSubscriptionQuery(name, host, port, dbname, user, password, publication string) string {
	return fmt.Sprintf("CREATE SUBSCRIPTION %s CONNECTION 'host=%s port=%s dbname=%s user=%s password=%s' PUBLICATION %s WITH (copy_data = false)", name, host, port, dbname, user, password, publication)
}

func DropSubscriptionQuery(name string) string {
	return fmt.Sprintf("DROP SUBSCRIPTION %s", name)
}
