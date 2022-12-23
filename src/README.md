<!-- This README file is going to be the one displayed on the Grafana.com website for your plugin -->

# Firestore

[Google Firestore](https://cloud.google.com/firestore) Data Source Plugin for [Grafana](https://grafana.com/).

Grafana Firestore Data Source Plugin enables integrating data on Firestore on to Grafana dashboards.

## Features
- [x] Use Google Firestore as data source for Grafana dashboards
- [x] Configure Firestore data source with GCP `Project Id` and [`Service Account`](https://cloud.google.com/firestore/docs/security/iam) for authentication
- [x] Store `Service Account` data source configuration in Grafana encrypted storage [Secure JSON Data](https://grafana.com/docs/grafana/latest/developers/plugins/add-authentication-for-data-source-plugins/#encrypt-data-source-configuration)
- [x] Query Firestore [collections](https://firebase.google.com/docs/firestore/data-model#collections) and path to collections
- [x] Auto detect data types: `string`, `number`, `boolean`, `json`
- [ ] Query selected fields from collection
- [ ] Order query results
- [ ] Limit query results
- [ ] Count query results
- [ ] Query [Collection Groups](https://firebase.blog/posts/2019/06/understanding-collection-group-queries)
- [ ] Use of [Grafafa global variables](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#global-variables) in queries.
