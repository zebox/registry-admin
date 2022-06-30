
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    BooleanField,
    EditButton,
    DeleteButton,
    ReferenceInput,
    SelectInput,
    TextInput
   
} from 'react-admin';


interface IRoleList {
    id: string;
    name: string;
}

export const RoleList: Array<IRoleList>=[
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
        filters={userFilters}
    >
        <Datagrid bulkActionButtons={false}>
                <TextField source="id" />
                <TextField source="login" />
                <TextField source="name" />
                <TextField source="role" />
                <BooleanField source="blocked" />
                <EditButton />
                <DeleteButton />
        </Datagrid>
    </List>
);


export default UserList;