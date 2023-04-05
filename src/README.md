<!-- This README file is going to be the one displayed on the Grafana.com website for your plugin -->

# Firestore

[Google Firestore](https://cloud.google.com/firestore) Data Source Plugin for [Grafana](https://grafana.com/).

Grafana Firestore Data Source Plugin enables integrating data on Firestore on to Grafana dashboards.

It uses [FireQL](https://github.com/pgollangi/FireQL) to capture user query that translated to issue queries on Firestore and construct results.

> [FireQL](https://github.com/pgollangi/FireQL) is a Go library to query Google Firestore database using SQL syntax.

## Features
- Use Google Firestore as a data source for Grafana dashboards
- Configure Firestore data source with GCP `Project Id` and [`Service Account`](https://cloud.google.com/firestore/docs/security/iam) for authentication
- Store `Service Account` data source configuration in Grafana encrypted storage [Secure JSON Data](https://grafana.com/docs/grafana/latest/developers/plugins/add-authentication-for-data-source-plugins/#encrypt-data-source-configuration)
- Query Firestore [collections](https://firebase.google.com/docs/firestore/data-model#collections) and path to collections
- Auto detect data types: `string`, `number`, `boolean`, `json`, `time.Time`
- Query selected fields from the collection
- Order query results
- Limit query results
- Query [Collection Groups](https://firebase.blog/posts/2019/06/understanding-collection-group-queries)


## Firestore data source configuration

![](https://raw.githubusercontent.com/pgollangi/firestore-grafana-datasource/main/src/screenshots/firestore-datasource-configuration.png)

## Using datasource 
![](https://raw.githubusercontent.com/pgollangi/firestore-grafana-datasource/main/src/screenshots/query-with-firestore-datasource.png)

## Migration instructions from version `<=0.1.0` to `0.2.0`

Starting from version `0.2.0` firestore datasource plugin started using [`FireQL`](https://github.com/pgollangi/FireQL), a SQL query language to query collections on Google Firestore. Util `0.1` version it was just collection ID that can configured in the query section, where as starting from `0.2.0`, it is query editor. 

For example, if you have collection `test_collection` configured in `v0.1`, you have to change it to `select * from test_collection` in `v0.2.0`.