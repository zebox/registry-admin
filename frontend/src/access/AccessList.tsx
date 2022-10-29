
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
    ReferenceField,
} from 'react-admin';
import { DisabledField } from "../components/DisabledField";



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
            <DisabledField source="disabled" />
            <EditButton alignIcon="right" />
            <DeleteButton alignIcon="right" />
        </Datagrid>
    </List>
);


export default AccessList;