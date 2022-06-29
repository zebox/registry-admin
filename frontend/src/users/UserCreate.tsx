
import { Create, PasswordInput, TextInput,TextField,ReferenceField, SelectInput, SimpleForm } from 'react-admin';

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
                    choices={[
                        { id: 'user', name: 'User' },
                        { id: 'manager', name: 'Manager' },
                        { id: 'admin', name: 'Admin' }
                    ]} />
            </SimpleForm>

        </Create>
    )
};

export default UserCreate;