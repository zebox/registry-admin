
import * as React from "react";
import { AutocompleteInput, BooleanInput, Edit, TextInput, SimpleForm, ReferenceInput, SelectInput,useInput, useTranslate, useRecordContext, useDataProvider } from 'react-admin';
import { ActionList } from "./AccessCreate";
import TextField from '@mui/material/TextField';
import Autocomplete from '@mui/material/Autocomplete';
import { repositoryBaseResource } from "../registry/RepositoryShow";


const AccessEdit = () => {
    const translate = useTranslate();
    return (
        <Edit title={translate('resources.groups.edit_title')}  >
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" />
                <ReferenceInput source="owner_id" reference="users">
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} />
                </ReferenceInput>
                {/*  <ReferenceInput source='resource_name' reference="registry/catalog">
                    <AutocompleteInput sx={{ width: "30%" }} optionText="repository_name" optionValue="repository_name" label={translate('resources.accesses.fields.resource_name')} />
                </ReferenceInput> */}
                <RepositoryAutocomplete source="resource_name" />
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />

                <SelectInput
                    label={translate('resources.accesses.fields.action')}
                    source="action"
                    choices={ActionList} />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Edit>
    )
};

interface RepositoryRecord {

    id: number;
    repository_name: string;
    tag: string;
    digest: string;
    size: string;
    pull_counter: number;
    timestamp: number;
    raw?: string

}

const RepositoryAutocomplete = (props:any) => {
    const {source} = props;
    const translate = useTranslate();
    const record = useRecordContext();
    const dataProvider = useDataProvider();
    const [option, setOptions] = React.useState<RepositoryRecord[] | never[] >([]);
    const {
        field,
        fieldState: { isTouched, error },
        formState: { isSubmitted }
    } = useInput(props);

    const fetchRepositoryData = (event: any): void => {
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


    return <Autocomplete
        sx={{ width: 300 }}
        onInputChange={(e) => fetchRepositoryData(e)}
        options={option}
        isOptionEqualToValue={(o,v)=>{return true}}
        getOptionLabel={(option: RepositoryRecord) => option.repository_name}
        id="clear-on-escape"
        renderInput={(params) => (
            <TextField {...params} {...field} value={record[source]} label={translate('resources.accesses.fields.resource_name')} variant="standard" />
        )}
    />
}

export default AccessEdit;