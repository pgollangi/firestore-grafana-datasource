import React, { ChangeEvent, PureComponent } from 'react';
import {
  QueryField
  // Form, InlineField, InlineFieldRow, Input
} from '@grafana/ui';
// import { FieldValues } from "react-hook-form"
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, FirestoreQuery } from '../types';

type Props = QueryEditorProps<DataSource, FirestoreQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  timeoutId: NodeJS.Timeout | undefined
  onCollectionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query });
    this.runQuery(onRunQuery)
  };

  onQueryChange = (newQuery: string) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, query: newQuery});
    this.runQuery(onRunQuery)
  };

  runQuery = (onRunQuery: () => void) => {
    if (this.timeoutId) {
      clearTimeout(this.timeoutId)
    }
    this.timeoutId = setTimeout(() => {
      onRunQuery();
      this.timeoutId = undefined
    }, 500)
  }

  onFormSubmit = () => {

  }

  render() {
    const {  query } = this.props.query;

    // const defaultValues: FieldValues = {
    //       where: [{ field: 'Janis', op: 'Joplin', value: "Va" }],
    // };

    return (
      <div>
        <div className="gf-form"> 
         <QueryField query={query} placeholder="FireQL query" portalOrigin="" onChange={this.onQueryChange}></QueryField>
        </div>
      </div>
    );
  }
}
