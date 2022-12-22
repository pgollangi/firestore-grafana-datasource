import React, { ChangeEvent, PureComponent } from 'react';
import { TextArea, InlineField, Input } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> { }

interface State { }

export class ConfigEditor extends PureComponent<Props, State> {
  onProjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      projectId: event.target.value,
    };
    onOptionsChange({
      ...options, jsonData
    });
  };

  // Secure field (only sent to the backend)
  onServiceAccountChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      serviceAccount: event.target.value,
    };

    onOptionsChange({
      ...options,
      jsonData
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData } = options;
    // const secureJsonData = (options.secureJsonData || {}) as FirestoreSecureJsonData;

    return (
      <div className="gf-form-group">
        <div className='gf-form-group'>
            <InlineField label="Project Id" labelWidth={20} tooltip="Project Id of on GCP console to fetch firestore resources.">
                <Input 
                  onChange={this.onProjectIdChange}
                  value={jsonData.projectId || ''}
                  placeholder="GCP Project ID" 
                  width={40}></Input>
              </InlineField>
              <InlineField label="Service Account" labelWidth={20} tooltip="Service Account having previliges to read all firestore resources. Least role expected is 'roles/datastore.viewer'">
                <TextArea
                  value={jsonData.serviceAccount || ''}
                  label="Service Account"
                  placeholder="Paste Service Account here"
                  onChange={this.onServiceAccountChange}
                  cols={80}
                  rows={10}
                />
              </InlineField>
        </div>

      </div>
    );
  }
}
