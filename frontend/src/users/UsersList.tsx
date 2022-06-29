
import * as React from "react";
import { List, Datagrid, TextField, BooleanField, EditButton,DeleteButton } from 'react-admin';

const UserList = () =>(
        <List 
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        >
            <Datagrid  bulkActionButtons={false}>
                <TextField source="id" />
                <TextField source="login" />
                <TextField source="name" />
                <TextField source="role" />
                <BooleanField source="blocked" />
                <EditButton />
                <DeleteButton/>
            </Datagrid>
        </List>    
);


export default UserList;