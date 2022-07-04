
import * as React from "react";
import {
    BooleanField,
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
    ReferenceField

} from 'react-admin';



const AccessList = () => (
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
    >
        <Datagrid bulkActionButtons={false} >
            <TextField source="id" />
            <TextField source="name" />
            <ReferenceField source="owner_id" reference="users">
                <TextField source="name" />
            </ReferenceField>
            <TextField source="type" />
            <TextField source="resource_name" />
            <TextField source="action" />
            <BooleanField source="disabled" />
            <EditButton alignIcon="right" />
            <DeleteButton alignIcon="right" />
        </Datagrid>
    </List>
);


export default AccessList;