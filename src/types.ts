import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface FirestoreQuery extends DataQuery {
  query: string
}

export const DEFAULT_QUERY: Partial<FirestoreQuery> = {
};

/**
 * These are options configured for each DataSource instance
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  projectId: string;
  serviceAccount: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface FirestoreSecureJsonData {
  serviceAccount: string;
}
