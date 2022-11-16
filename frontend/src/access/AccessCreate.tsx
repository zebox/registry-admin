
import {required, AutocompleteInput, Create, TextInput, SimpleForm, ReferenceInput, BooleanInput, useTranslate } from 'react-admin';
import { RepositoryAutocomplete } from '../components/RepositoryAutocompleteField';
import RepositoryAction from "../components/RepositoryAction";

interface IActionList {
    id: string;
    name: string;
};

export const ActionList: Array<IActionList> = [
    { id: 'push', name: 'push' },
    { id: 'pull', name: 'pull' }
];



export const AccessCreate = () => {
    const translate = useTranslate();

    const onResetHandler = (event: any) => {
        console.log(event);
    }
    return (
        <Create title={translate('resources.accesses.add_title')} >
            <SimpleForm>
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" validate={required()}/>
                <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')} >
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" onReset={onResetHandler} label={translate('resources.accesses.fields.owner_id')}  validate={required()}/>
                </ReferenceInput>
                <RepositoryAutocomplete source="resource_name"/>
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />

                <RepositoryAction source="action" validate={required()}/>
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Create>
    )
};

export default AccessCreate;