
import {
    List,
    Datagrid,
    TextField,
    BooleanField,
    EditButton,
    DeleteButton,
    SelectInput,
    TextInput,
    ReferenceField

} from 'react-admin';

import { SearchFieldTranslated } from '../helpers/Helpers'

interface IRoleList {
    id: string;
    name: string;
}

export const RoleList: Array<IRoleList> = [
    { id: 'user', name: 'User' },
    { id: 'manager', name: 'Manager' },
    { id: 'admin', name: 'Admin' }
]


const userFilters = [
    <TextInput source="q" label="Search" alwaysOn />,
    <SelectInput
        source="role"
        defaultValue={"user"}
        emptyValue={null}
        choices={RoleList} />
];


const UserList = () => (
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated([<SelectInput
            source="role"
            defaultValue={"user"}
            emptyValue={null}
            choices={RoleList} />])}
    >
        <Datagrid bulkActionButtons={false}>
            <TextField source="id" />
            <TextField source="login" />
            <TextField source="name" />
            <ReferenceField source="group" reference="groups">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="role" />
            <BooleanField source="blocked" />
            <EditButton />
            <DeleteButton />
        </Datagrid>
    </List>
);


export default UserList;