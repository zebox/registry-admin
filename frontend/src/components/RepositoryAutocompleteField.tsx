import React, { useState } from "react";
import { Identifier, useDataProvider, useInput, useRecordContext, useTranslate } from "react-admin";
import { repositoryBaseResource } from "../registry/RepositoryShow";
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';

interface RepositoryRecord {

    id?: Identifier;
    repository_name: string;
    tag?: string;
    digest?: string;
    size?: string;
    pull_counter?: number;
    timestamp?: number;
    raw?: string

}

export const RepositoryAutocomplete = (props: any) => {

    const source_name = 'repository_name';
    const translate = useTranslate();
    const record = useRecordContext();

    var defaultValue: RepositoryRecord = {
        repository_name: record ? record[source_name] : null
    };

    const dataProvider = useDataProvider();
    const [repoSelectValue, setRepoSelectValue] = React.useState<string | null>(record ? record[source_name] : null);
    const [repoInputValue, setRepoInputValue] = useState('');
    const [options, setOptions] = useState<RepositoryRecord[] | never[]>([]);
    const {
        field,
        fieldState: {isTouched, error},
        formState: {isSubmitted}
    } = useInput(props);


    const fetchRepositoryData = (event: any, newInputValue: string): void => {

        setRepoInputValue(newInputValue);
        if (!event || event == null) {
            return;
        }

        dataProvider.getList(
            repositoryBaseResource,
            {
                pagination: {page: 1, perPage: 20},
                sort: {field: 'repository_name', order: 'ASC'},
                filter: {q: newInputValue}
            }
        ).then(({ data, total }) => {
            if (total && total > 0) {
                setOptions(data);
            }
        })
    };

    return <Autocomplete
        sx={{width: 300}}
        onInputChange={fetchRepositoryData}
        onChange={async (event: any, newValue: string | null) => {
            setRepoSelectValue(newValue);
            if (newValue === null) {
                return;
            }
            await setRepoInputValue(newValue);
            field.onBlur();
            record[source_name] = newValue;
            console.log(newValue);
        }}
        options={options.map((item: RepositoryRecord) => item.repository_name)}
        value={repoSelectValue}
        inputValue={repoInputValue}
        isOptionEqualToValue={(o, v) => {
            return true
        }}
        id="repository-autocomplete-search"
        renderInput={(params) => (
            <TextField
                {...params}
                {...field}
                label={translate('resources.accesses.fields.resource_name')}
                variant="standard"/>
        )}
    />
}


