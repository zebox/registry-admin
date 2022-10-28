
import * as React from "react";
import {
    BooleanField,
    List,
    Datagrid,
    TextField,
    EditButton,
    DeleteButton,
    ReferenceField,
    useRecordContext

} from 'react-admin';
import DoDisturbIcon from '@mui/icons-material/DoDisturb';

const DisabledField = ({ source }: any) => {
    const record = useRecordContext();
    return (
        <>
            {record[source] ? <DoDisturbIcon sx={{color:"red"}}/> : ""}
        </>
    )
}
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