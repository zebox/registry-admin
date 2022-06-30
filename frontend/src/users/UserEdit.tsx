
import * as React from "react";
import { Edit, TextField, PasswordInput, TextInput, BooleanInput, ReferenceInput , SelectInput, SimpleForm,Labeled, useTranslate } from 'react-admin';

const UserEdit = () => {
    const translate=useTranslate();
    return (
        <Edit title={translate('resources.users.edit_title')}  >
            <SimpleForm >
                <TextInput label="ID" source="id" disabled />
                <TextInput label={translate('resources.users.fields.login')} source="login" disabled />
                <TextInput label={translate('resources.users.fields.name')} source="name" autoComplete='off' />
                <PasswordInput label={translate('resources.users.fields.password')} source="password" autoComplete="new-password" />
                <ReferenceInput  label={translate('resources.users.fields.group')} source="group" reference="groups">
                   <SelectInput  emptyValue={null} emptyText='' optionText="name" optionValue="id"/>
                </ReferenceInput >
                <SelectInput
                    label={translate('resources.users.fields.role')}
                    source="role"
                    defaultValue={"user"}
                    emptyValue={null}
                    choices={[
                        { id: 'user', name: 'User' },
                        { id: 'manager', name: 'Manager' },
                        { id: 'admin', name: 'Admin' }
                    ]} />
                <BooleanInput label={translate('resources.users.fields.blocked')} source="blocked" />
                <TextInput label={translate('resources.users.fields.description')} source="description" autoComplete='off' fullWidth />
            </SimpleForm>
        </Edit>
    )
};

export default UserEdit;