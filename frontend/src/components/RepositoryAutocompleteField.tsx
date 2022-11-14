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

    const { source } = props;
    const translate = useTranslate();
    const record = useRecordContext();

    var defaultValue: RepositoryRecord = {
        repository_name: record ? record[source] : ""
  };

    const dataProvider = useDataProvider();
    const [repo, setRepo] = useState<RepositoryRecord | undefined>(defaultValue);
    const [option, setOptions] = useState<RepositoryRecord[] | never[]>([]);
    const {
        field,
        fieldState: { isTouched, error },
        formState: { isSubmitted }
    } = useInput(props);

    
   

    const fetchRepositoryData = (event: any): void => {
        if (!event || event == null) {
            return;
        }

        const searchValue = event.target.value;
        dataProvider.getList(
            repositoryBaseResource,
            {
                pagination: { page: 1, perPage: 20 },
                sort: { field: 'repository_name', order: 'DESC' },
                filter: { q: searchValue }
            }
        ).then(({ data, total }) => {
            if (total && total > 0) {
                setOptions(data);
            }
        })
    };

    const handleValueChange=(e:any)=>{
        console.log(e.target.value);
        // defaultValue.repository_name=e.target.value;
        if (e==null) {
            return;
        }
        setRepo(e.target.value);

    }

    return <Autocomplete
        sx={{ width: 300 }}
        onInputChange={(e) => fetchRepositoryData(e)}
        options={option}
        //defaultValue={defaultValue}
        value={repo}
       
        isOptionEqualToValue={(o, v) => { return true }}
        getOptionLabel={(option: RepositoryRecord) => option.repository_name}
        id="repository-autocomplete-search"
        renderInput={(params) => (
            <TextField 
            {...params} 
            {...field} 
            onChange={(e)=>{handleValueChange(e)}}
            label={translate('resources.accesses.fields.resource_name')}
            variant="standard" />
        )}
    />
}


