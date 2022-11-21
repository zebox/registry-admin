import * as React from "react";
import {Edit, TextInput, SimpleForm, useTranslate, usePermissions, NotFound} from 'react-admin';
import {requirePermission} from "../helpers/Helpers";

const UserEdit = () => {
    const translate = useTranslate();
    const {permissions} = usePermissions();
    return (requirePermission(permissions, 'admin') ?
            <Edit title={translate('resources.groups.edit_title')}>
                <SimpleForm>
                    <TextInput label="ID" source="id" disabled/>
                    <TextInput label={translate('resources.groups.fields.name')} source="name" autoComplete='off'/>
                    <TextInput label={translate('resources.groups.fields.description')} source="description"
                               autoComplete='off' fullWidth/>
                </SimpleForm>
            </Edit> : <NotFound/>
    )
};

export default UserEdit;