
import * as React from "react";
import { Edit, TextInput, SimpleForm, useTranslate } from 'react-admin';

const UserEdit = () => {
    const translate = useTranslate();
    return (
        <Edit title={translate('resources.groups.edit_title')}  >
            <SimpleForm >
                <TextInput label="ID" source="id" disabled />
                <TextInput label={translate('resources.groups.fields.name')} source="name" autoComplete='off' />
                <TextInput label={translate('resources.groups.fields.description')} source="description" autoComplete='off' fullWidth />
            </SimpleForm>
        </Edit>
    )
};

export default UserEdit;