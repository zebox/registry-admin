
import * as React from "react";
import { BooleanInput, Edit, TextInput, SimpleForm, ReferenceInput, SelectInput, useTranslate } from 'react-admin';
import {ActionList} from "./AccessCreate";

const AccessEdit = () => {
    const translate = useTranslate();
    return (
        <Edit title={translate('resources.groups.edit_title')}  >
            <SimpleForm >
                <TextInput label={translate('resources.accesses.fields.name')} source="name" />
                <ReferenceInput  source="id" reference="users">
                    <SelectInput label={translate('resources.accesses.fields.owner_id')} source="owner_id" emptyValue={null} emptyText='' optionText="name" optionValue="id" />
                </ReferenceInput>
                <TextInput label={translate('resources.accesses.fields.resource_type')} source="type" />
                <TextInput label={translate('resources.accesses.fields.resource_name')} source="resource_name" />
                <SelectInput
                    label={translate('resources.accesses.fields.action')}
                    source="action"
                    defaultValue={"pull"}
                    emptyValue={null}
                    choices={ActionList} />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>
        </Edit>
    )
};

export default AccessEdit;