
import {required, AutocompleteInput, Create, TextInput, SimpleForm, ReferenceInput, BooleanInput, useTranslate } from 'react-admin';
import { RepositoryAutocomplete } from '../components/RepositoryAutocompleteField';
import RepositoryAction from "../components/RepositoryAction";

export const AccessCreate = (props:any) => {
    const { source, ...rest } = props;
    const translate = useTranslate();
  
    return (
        <Create title={translate('resources.accesses.add_title')} >
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" validate={required()}/>
                <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')} >
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" label={translate('resources.accesses.fields.owner_id')}  validate={required()}/>
                </ReferenceInput>
                <RepositoryAutocomplete source="resource_name" {...rest} validate={required()} />            
                <RepositoryAction source="action" validate={required()} defaultValue="pull"/>
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Create>
    )
};

export default AccessCreate;