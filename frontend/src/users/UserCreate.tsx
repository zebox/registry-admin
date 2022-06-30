
import { Create, PasswordInput, TextInput,TextField,ReferenceField, SelectInput, SimpleForm } from 'react-admin';
import { RoleList } from "./UsersList";

export const UserCreate = () => {

    return (
        <Create title="Add User" >
            <SimpleForm>
                <TextInput source="login" autoComplete="new-login" />
                <TextInput source="name" autoComplete="new-name" />
                <PasswordInput source="password" autoComplete="new-password" />
                <ReferenceField source="group" reference="groups">
                    <TextField source="id" />
                </ReferenceField>
                <SelectInput
                    source="role"
                    defaultValue={"user"}
                    emptyValue={null}
                    emptyText=''
                    choices={RoleList} />
            </SimpleForm>

        </Create>
    )
};

export default UserCreate;