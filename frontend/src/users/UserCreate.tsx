
import { BooleanInput, Create, PasswordInput, TextInput, ReferenceInput, SelectInput, SimpleForm, useTranslate } from 'react-admin';
import { RoleList } from "./UsersList";

export const UserCreate = () => {
    const translate = useTranslate();
    return (
        <Create title="Add User" >
            <SimpleForm>
                <TextInput label={translate('resources.users.fields.login')} source="login" disabled />
                <TextInput label={translate('resources.users.fields.name')} source="name" autoComplete='off' />
                <PasswordInput label={translate('resources.users.fields.password')} source="password" autoComplete="new-password" />
                <ReferenceInput source="group" reference="groups">
                    <SelectInput label={translate('resources.users.fields.group')} emptyValue={null} emptyText='' optionText="name" optionValue="id" />
                </ReferenceInput >
                <SelectInput
                    label={translate('resources.users.fields.role')}
                    source="role"
                    defaultValue={"user"}
                    emptyValue={null}
                    choices={RoleList} />
                <BooleanInput label={translate('resources.users.fields.blocked')} source="blocked" />
                <TextInput label={translate('resources.users.fields.description')} source="description" autoComplete='off' fullWidth />
            </SimpleForm>
        </Create>
    )
};

export default UserCreate;