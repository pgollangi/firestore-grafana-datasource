import React, { ChangeEvent, PureComponent } from 'react';
import {
  Form, InlineField, InlineFieldRow, Input
} from '@grafana/ui';
import { FieldValues } from "react-hook-form"
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, FirestoreQuery } from '../types';

type Props = QueryEditorProps<DataSource, FirestoreQuery, MyDataSourceOptions>;

export class QueryEditor extends PureComponent<Props> {
  timeoutId: NodeJS.Timeout | undefined
  onCollectionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, collectionPath: event.target.value, limit: 0 });
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
    const { collectionPath } = this.props.query;

    const defaultValues: FieldValues = {
          where: [{ field: 'Janis', op: 'Joplin', value: "Va" }],
    };

    return (
      <div>
        <div className="gf-form"> 

          <Form onSubmit={this.onFormSubmit} style={{ maxWidth: '100%' }} defaultValues={defaultValues}
          // onChange={this.onFormChange}
          >
            {({ control, register }) => {
              return (
                <div style={{ width: '100%' }}>
                  <InlineFieldRow>
                    <InlineField label="Collection" tooltip="Enter collection id or path to a collection to query a single collection." required >
                      <Input value={collectionPath} onChangeCapture={this.onCollectionChange} placeholder="Path to collection" width={40}/>
                    </InlineField>
                    {/* <InlineField label="Count" >
                      <InlineSwitch id="count" />
                    </InlineField> */}
                  </InlineFieldRow>

                  {/* <InlineFieldRow>
                    <FieldArray control={control} name="where">
                      {({ fields, append }) => {
                        return (
                          <div style={{ marginBottom: '1rem' }}>
                            {fields.map((field, index) => (
                            <InlineFieldRow >
                              <InlineField label="Field" >
                                <Input id="field" placeholder="Field" />
                              </InlineField>
                              <InlineField label="Operator" >
                                <Select
                                width={16}
                                onChange={console.log}
                                options={[
                                  { value: "==", label: "==", description: "Equal to" },
                                  { value: "!=", label: "!=", description: "Not Equal to" },
                                ]}
                                />
                              </InlineField>
                              <InlineField label="Value" >
                                <Input id="value" placeholder="Value" />
                              </InlineField>
                              <HorizontalGroup align="center" width='200'>
                              <IconButton  name="plus-circle" onClick={() => append({ field: 'Roger', value: 'Waters' })}></IconButton>
                              <IconButton  name="minus-circle" onClick={() => fields.splice(fields.indexOf(field), 1)}  hidden={fields.length <= 1}></IconButton>
                              </HorizontalGroup>
                            </InlineFieldRow>
                            ))}
                          </div>
                        )
                      }}
                    </FieldArray>
                  </InlineFieldRow>  */}
                  {/* <InlineFieldRow>
                    <InlineField label="Order By" >
                      <Input id="orderBy" placeholder="Id" />
                    </InlineField>
                  </InlineFieldRow> */}

                </div>

              )
            }}
          </Form>
        </div>
      </div>
    );
  }
}
