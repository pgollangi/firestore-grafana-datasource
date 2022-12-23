import React, { ChangeEvent, PureComponent } from 'react';
import { InlineField, Input, SecretTextArea } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { FirestoreSecureJsonData, MyDataSourceOptions } from '../types';

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> { }

interface State { }

export class ConfigEditor extends PureComponent<Props, State> {
  onProjectIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      projectId: event.target.value.trim(),
    };
    onOptionsChange({
      ...options, jsonData
    });
  };

  // Secure field (only sent to the backend)
  onServiceAccountChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        serviceAccount: event.target.value,
      },
    });
  };

  onResetServiceAccount = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        serviceAccount: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        serviceAccount: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as FirestoreSecureJsonData;

    return (
      <div className="gf-form-group">
        <div className='gf-form-group'>
          <InlineField required label="Project Id" labelWidth={20}
            tooltip="Project Id of on GCP console to fetch firestore resources.">
            <Input
              onChange={this.onProjectIdChange}
              value={jsonData.projectId || ''}
              placeholder="Unique identifier for the GCP Project"
              width={40}></Input>
          </InlineField>
          <InlineField required label="Service Account" labelWidth={20}
            tooltip="Service Account having previliges to read all firestore resources. Least role expected is 'roles/datastore.viewer'">
            <SecretTextArea
              label="Service Account"
              placeholder={`{
      ...
      "private_key_id": "...",
      "private_key": "...",
      "client_email": "...",
      "client_id": "..."
      ...
}
              `}

              value={secureJsonData.serviceAccount || ''}
              isConfigured={(secureJsonFields && secureJsonFields.serviceAccount) as boolean}
              onReset={this.onResetServiceAccount}
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
