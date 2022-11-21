import * as React from "react";
import {
    Edit,
    PasswordInput,
    TextInput,
    BooleanInput,
    ReferenceInput,
    SelectInput,
    SimpleForm,
    useTranslate,
    usePermissions,
    NotFound
} from 'react-admin';
import {requirePermission} from "../helpers/Helpers";
import {RoleList} from "./UsersList";

const UserEdit = () => {
    const translate = useTranslate();
    const {permissions} = usePermissions();
    return (requirePermission(permissions, 'admin') ?
            <Edit title={translate('resources.users.edit_title')}>
                <SimpleForm>
                    <TextInput label="ID" source="id" disabled/>
                    <TextInput label={translate('resources.users.fields.login')} source="login" disabled/>
                    <TextInput label={translate('resources.users.fields.name')} source="name" autoComplete='off'/>
                    <PasswordInput label={translate('resources.users.fields.password')} source="password"
                                   autoComplete="new-password"/>
                    <ReferenceInput source="group" reference="groups">
                        <SelectInput label={translate('resources.users.fields.group')} emptyValue={''} emptyText=''
                                     optionText="name" optionValue="id"/>
                    </ReferenceInput>
                    <SelectInput
                        label={translate('resources.users.fields.role')}
                        source="role"
                        defaultValue={"user"}
                        emptyValue={''}
                        choices={RoleList}/>
                    <BooleanInput label={translate('resources.users.fields.blocked')} source="blocked"/>
                    <TextInput label={translate('resources.users.fields.description')} source="description"
                               autoComplete='off' fullWidth/>
                </SimpleForm>
            </Edit> : <NotFound/>
    )
};

export default UserEdit;