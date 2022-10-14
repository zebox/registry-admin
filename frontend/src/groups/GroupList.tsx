
import * as React from "react";
import {
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,

} from 'react-admin';


const GroupList = () => (
    <List
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
    >
        <Datagrid bulkActionButtons={false}

         >
                <TextField source="id"  />
                <TextField source="name" />
                <EditButton alignIcon="right" />
                <DeleteButton alignIcon="right" />
        </Datagrid>
    </List>
);


export default GroupList;