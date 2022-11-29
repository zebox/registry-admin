import { TranslationMessages } from 'react-admin';
import englishMessages from 'ra-language-english';

const customEnglishMessages: TranslationMessages = {
    ...englishMessages,
    portal: {
        configuration: "Settings",
        language: "Language",
        theme: {
            type: "Theme type",
            light: "Light",
            dark: "Dark"
        }
    },
    resources: {
        commands: {
            users_name: "Users",
            groups_name: "Groups",
            access_name: "Accesses",
            repository_name: "Repositories"
        },
        users: {
            name: "Users",
            add_title: "Add new user",
            edit_title: "Edit user entry",
            fields: {
                login: "Login",
                name: "Username",
                password: "Password",
                group: "Group",
                role: "Role",
                blocked: "User blocked",
                description: "Description"

            }
        },
        groups: {
            name: "Groups",
            add_title: "Add new group",
            edit_title: "Edit group",
            fields: {
                name: "Username",
                description: "Description"

            }
        },
        accesses:{
            name: "Accesses",
            add_title: "Add new access",
            edit_title: "Edit access",
            fields: {
                name: "Access name",
                owner_id: "User",
                resource_type: "Resource type",
                resource_name: "Repository name",
                action:"Allowed action",
                disabled:"Disabled"
            },
            messages:{
                access_tooltip:"show repositories access"
            }
        },
        repository:{
            fields:{
                name: "Name",
                size: "Size",
                tag:"Tag",
                date:"Date",
                digest:"Digest",
                details:"DETAILS"
            },
            title:"Repository details",
            tag_list_title:"Tags list",
            pull_counter:"Pull counter: ",
            tag_digest:"Digest: ",
            tag_media_type:"Type: ",
            image_platform_details:"Platform data",
            image_config_details:"Configuration",
            image_history_details:"Histroy",
            message_empty_page:"Repositories entry not found.",
            message_config_data_not_loading:"Config data not loading",
            message_sync_about:"Run synchronization repositories data with RA and Docker Registry",
            message_sync_repo:"Synchronize repositories from registry",
            message_syncing_repo:"Repositories sync in progress...",
            message_error_syncing_repo:"Synchronization error",
            message_copied_to_clipboard: "Copied to clipboard",
            message_repo_syncing_running:"Synchronization currently running. Please wait for complete previous task.",
            button_sync:"Sync"

        }
    }
};

export default customEnglishMessages;
