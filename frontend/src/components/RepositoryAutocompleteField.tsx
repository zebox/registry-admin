import React, { useEffect, useState } from "react";
import { Identifier, useDataProvider, useInput, useRecordContext, useTranslate } from "react-admin";
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
    const [repoSelectValue, setRepoSelectValue] = React.useState<string | null>(record ? record[source_name] : undefined);
    const [options, setOptions] = useState<readonly RepositoryRecord[] | never[]>([]);
    const { onChange, onBlur, ...rest } = props;
    const {
        field,
        fieldState: { isTouched, error },
        formState: { isSubmitted },
        isRequired
    } = useInput({
        onChange,
        onBlur,
        ...props,
    });



    useEffect(() => {
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
        getRepositoryData();

        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, []);

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
                        error={(isTouched || isSubmitted) && error}
                        label={translate('resources.accesses.fields.resource_name')}
                        helperText={(isTouched || isSubmitted) && error ? error.message : ''}
                        required={isRequired}
                        {...rest}
                    />
                )}
            />
        )}
    />
}


