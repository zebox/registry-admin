import React, { useEffect, useState } from "react";
import { Identifier,  useDataProvider, useInput, useRecordContext, useTranslate, required } from "react-admin";
import { repositoryBaseResource } from "../registry/RepositoryShow";
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import { useForm, Controller } from "react-hook-form";
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
    const { control } = useForm();
    const source_name = 'resource_name';
    const translate = useTranslate();
    const record = useRecordContext();

    var defaultValue: RepositoryRecord = {
        repository_name: record ? record[source_name] : null
    };

    const dataProvider = useDataProvider();
    const [repoSelectValue, setRepoSelectValue] = React.useState<string | null>(record ? record[source_name] : null);
    const [repoInputValue, setRepoInputValue] = useState('');
    const [options, setOptions] = useState<readonly RepositoryRecord[] | never[]>([]);
     const {
         field,
         fieldState: { isTouched, error },
         formState: { isSubmitted }    
     } = useInput(props);


    useEffect(() => {
        getRepositoryData();
    }, []);


    const fetchOptionsData = (event: any, newInputValue: string): void => {
        setRepoInputValue(newInputValue);
        if (!event || event == null) {
            return;
        }
        getRepositoryData(newInputValue);
    };

    const getRepositoryData = (searchValue: string | void) => {
        dataProvider.getList(
            repositoryBaseResource,
            {
                pagination: { page: 1, perPage: 20 },
                sort: { field: 'repository_name', order: 'ASC' },
                filter: { q: searchValue !== "" ? searchValue : "" }
            }
        ).then(({ data, total }) => {
            if (total && total > 0) {
                setOptions(data);
            }
        })
    }
    return <Controller
        control={control}
        defaultValue={defaultValue.repository_name}
        name={source_name}
        render={() => (
            <Autocomplete
                sx={{ width: 300 }}
                value={repoSelectValue}
                options={options.map((item: RepositoryRecord) => item.repository_name)}
                onChange={(_, data: any) => {
                    setRepoSelectValue(data);
                    field.onChange(data);
                }}
                renderInput={(params) => (
                    <TextField
                        {...params}
                        {...field}
                        inputRef={field.ref}
                        label={translate('resources.accesses.fields.resource_name')}
                    />
                )}
            />
        )}
    />
    /*  <Autocomplete
         sx={{ width: 300 }}
         onInputChange={fetchOptionsData}
        onChange={async (_: any, newValue: string | null) => {
              setRepoSelectValue(newValue);
              if (newValue === null) {
                  return;
              }
              setRepoInputValue(newValue);
              field.onChange(newValue);
              record[source_name] = newValue;
              console.log(newValue);
          }}  
         //{...field}
 
         options={options.map((item: RepositoryRecord) => item.repository_name)}
         value={repoSelectValue}
         inputValue={repoInputValue}
         isOptionEqualToValue={(o, v) => {
             record[source_name] = v;
             setRepoSelectValue(v);
             return true
         }}
         renderInput={(params) => (
             <TextField
                 {...params}
                 {...field}
                 //inputRef={field.ref}
                 label={translate('resources.accesses.fields.resource_name')}
                 variant="standard" />
         )}
     /> */

}


