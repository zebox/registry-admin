
import { AutocompleteInput, Create, TextInput, SimpleForm, SelectArrayInput, ReferenceInput, BooleanInput, SelectInput, useTranslate, TextField } from 'react-admin';
import { RepositoryAutocomplete } from '../components/RepositoryAutocompleteField';


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
                <TextInput sx={{ width: "30%" }} label={translate('resources.accesses.fields.name')} source="name" />
                <ReferenceInput source="owner_id" reference="users" label={translate('resources.accesses.fields.owner_id')}>
                    <AutocompleteInput sx={{ width: "30%" }} optionText="name" optionValue="id" onReset={onResetHandler} />
                </ReferenceInput>
{/*                 <ReferenceInput source='resource_name' reference="registry/catalog">
                    <AutocompleteInput sx={{ width: "30%" }} isOptionEqualToValue={(o,v)=>{return o===v}} optionText="repository_name"  optionValue="repository_name" label="Repository list" />
                </ReferenceInput> */}
                <RepositoryAutocomplete source="resource_name" />
                <TextInput
                    label={translate('resources.accesses.fields.resource_type')}
                    source="type"
                    defaultValue={"repository"}
                    disabled
                />

                <SelectInput
                    label={translate('resources.accesses.fields.action')}
                    source="action"
                    choices={ActionList} />
                <BooleanInput label={translate('resources.accesses.fields.disabled')} source="disabled" />
            </SimpleForm>

        </Create>
    )
};

export default AccessCreate;