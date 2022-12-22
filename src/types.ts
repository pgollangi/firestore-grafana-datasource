import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface FirestoreQueryCondition {
  path: string
  operator: string
  value: string
}

export interface FirestoreQueryOrderBy{
  path:  string
  direction: number // 1 for ASCENDING, 2 for DESCENDING
}

export interface FirestoreQuery extends DataQuery {
  collectionPath: string;
  select?: string[]
  where?: FirestoreQueryCondition[]
  orderBy?:  FirestoreQueryOrderBy[]
  limit?: number
  isCount?: boolean
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
