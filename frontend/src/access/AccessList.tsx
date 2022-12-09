import React, {useEffect, useState} from "react";
import {
    List,
    Datagrid,
    ReferenceInput,
    AutocompleteInput,
    TextField,
    EditButton,
    ReferenceField,
    SelectInput,
    useTranslate,
    usePermissions,
    useRecordContext
} from 'react-admin';
import { DeleteCustomButtonWithConfirmation } from '../components/DeleteCustomButtonWithConfirmation';
import {DisabledField} from "../components/DisabledField";
import {SearchFieldTranslated, requirePermission} from '../helpers/Helpers'


const ActionList: Array<IActionList> = [
    { id: 'push', name: 'push' },
    { id: 'pull', name: 'pull' }
];


const AccessList = (props:any) => {
  const translate = useTranslate();
  const { permissions } = usePermissions();

  return  <List
        hasCreate={requirePermission(permissions,'admin')}
        sort={{ field: 'name', order: 'ASC' }}
        perPage={25}
        filters={SearchFieldTranslated(translate,[
        <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
            <AutocompleteInput  optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')} />
        </ReferenceInput>,
        <SelectInput
        label={translate('resources.accesses.fields.action')}
        source="action"
        choices={ActionList} />
    ])}
    >
        <Datagrid bulkActionButtons={false} >
            <TextField source="name" label={translate('resources.accesses.fields.name')}/>
            <CustomUsersReference label={translate('resources.accesses.fields.owner_id')}/>
            <TextField source="type" label={translate('resources.accesses.fields.resource_type')}/>
            <TextField source="resource_name" label={translate('resources.accesses.fields.resource_name')} />
            <TextField source="action" label={translate('resources.accesses.fields.action')}/>
            <DisabledField source="disabled" label={translate('resources.accesses.fields.disabled')}/>
            {requirePermission(permissions, 'admin') ? <>
                <EditButton alignIcon="right"/>
                <DeleteCustomButtonWithConfirmation source="name" {...props}/>
            </> : null}
        </Datagrid>
  </List>
};


const CustomUsersReference = (props: any) => {
    const translate = useTranslate();
    const record = useRecordContext();
    const [isUserReference, setIsUserREference] = useState(false);

    const specsPermissionRefs = [
        {name: translate('resources.accesses.labels.label_for_all_users')},
        {name: translate('resources.accesses.labels.label_for_registered_users')}
    ];

    useEffect(() => {
        if (record && record.owner_id < 0) {
            setIsUserREference(true)
            console.log(record)
        }
    }, [record])

    return (
        !isUserReference ?
            <ReferenceField source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
                <TextField source="name"/>
            </ReferenceField> :
            <>{record.owner_id === -1000 ? specsPermissionRefs[0].name : specsPermissionRefs[1].name}</>
    )
}

interface IActionList {
    id: string;
    name: string;
};


export default AccessList;

