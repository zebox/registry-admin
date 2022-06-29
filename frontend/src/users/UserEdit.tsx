
import * as React from "react";
import { Edit, TextField, PasswordInput, TextInput, BooleanInput, ReferenceInput , SelectInput, SimpleForm } from 'react-admin';

const UserEdit = () => {

    return (
        <Edit title="Edit user" >
            <SimpleForm >
                <TextField label="ID" source="id" />
                <TextField label="Login" source="login" />
                <TextInput source="name" autoComplete='off' />
                <PasswordInput source="password" autoComplete="new-password" />
                <ReferenceInput  source="group" reference="groups">
                   <SelectInput  emptyValue={null} emptyText='' optionText="name" optionValue="id"/>
                </ReferenceInput >
                <SelectInput
                    source="role"
                    defaultValue={"user"}
                    emptyValue={null}
                    choices={[
                        { id: 'user', name: 'User' },
                        { id: 'manager', name: 'Manager' },
                        { id: 'admin', name: 'Admin' }
                    ]} />
                <BooleanInput source="blocked" />
                <TextInput source="description" autoComplete='off' fullWidth />
            </SimpleForm>
        </Edit>
    )
};

export default UserEdit;