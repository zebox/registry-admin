
import * as React from "react";
import { AutocompleteInput, BooleanInput, Edit, TextInput, SimpleForm, ReferenceInput, SelectInput, useTranslate, required  } from 'react-admin';
import { ActionList } from "./AccessCreate";
import { RepositoryAutocomplete } from "../components/RepositoryAutocompleteField";
import { repositoryBaseResource } from "../registry/RepositoryShow";


const AccessEdit = () => {

    const translate = useTranslate();
   

    return (
        <Edit title={translate('resources.groups.edit_title')}  >
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" validate={required()} />
                <ReferenceInput source="owner_id" reference="users">
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} validate={required()}/>
                </ReferenceInput>
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

export default AccessEdit;