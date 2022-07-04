
import { Create, TextInput, SimpleForm } from 'react-admin';
export const UserCreate = () => {

    return (
        <Create title="Add User" >
            <SimpleForm>
                <TextInput source="name" autoComplete="new-name" />
                <TextInput source="description" autoComplete='off' fullWidth />
            </SimpleForm>

        </Create>
    )
};

export default UserCreate;