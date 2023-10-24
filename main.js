const weekNames = {
    "9 – 15 Oct 2023": "Week 3",
    "16 – 22 Oct 2023": "Week 4",
    "23 – 29 Oct 2023": "Week 5",
    "30 Oct – 5 Nov 2023": "Consolidation Week"
};

document.addEventListener('DOMContentLoaded', function () {
    var calendarEl = document.getElementById('calendar');
    var calendar = new FullCalendar.Calendar(calendarEl, {
        headerToolbar: {
            left: '',
            center: 'title',
            right: 'prev,next today',
        },
        navLinks: false,
        nowIndicator: true,
        editable: true,
        initialView: 'timeGridWeek',
        locale: "en-GB",
        firstDay: 1,
        allDaySlot: false,
        eventClick: function (info) {
            document.getElementById('eventTitleView').innerText = info.event.title;
            document.getElementById('eventStartTimeView').textContent = info.event.start;
            document.getElementById('eventEndTimeView').textContent = info.event.end;

            checkUserPermissionToDelete().then((allowed) => {
                const deleteButton = document.getElementById('deleteEvent');
                if (allowed) {
                    deleteButton.style.display = 'block';
                } else {
                    deleteButton.style.display = 'none';
                }
            });

            $('#eventDetailsModal').modal('show');
        },
        events: "/get"
    });
    calendar.render();

    document.getElementById("week-name").innerText = weekNames[document.getElementById("fc-dom-1").innerText];
    document.getElementById("fc-dom-1").addEventListener("DOMCharacterDataModified", function () {
        document.getElementById("week-name").innerText = weekNames[document.getElementById("fc-dom-1").innerText];
    }, false);
});

// Fetch event types from an API endpoint
function fetchEventTypes() {
    // Simulate fetching event types
    return new Promise((resolve) => {
        setTimeout(() => {
            resolve(['Training', 'Recording', 'Engineering', 'Meeting', 'Other']);
        }, 1000);
    });
}

// Populate the event type dropdown
async function populateEventTypes() {
    const eventTypes = await fetchEventTypes();
    const eventTypeDropdown = document.getElementById('eventType');

    eventTypes.forEach((type) => {
        const option = document.createElement('option');
        option.value = type;
        option.text = type;
        eventTypeDropdown.appendChild(option);
    });
}

populateEventTypes();

document.getElementById("create-button").onclick = () => {
    const titleGroup = document.getElementById('titleGroup');
    titleGroup.style.display = ['Engineering', 'Meeting', 'Other'].includes(document.getElementById("eventType").value) ? 'block' : 'none';
}

// Show or hide the event title input based on event type
document.getElementById('eventType').addEventListener('change', function () {
    const selectedType = this.value;
    const titleGroup = document.getElementById('titleGroup');
    titleGroup.style.display = ['Engineering', 'Meeting', 'Other'].includes(selectedType) ? 'block' : 'none';
});

// Handle event submission
document.getElementById('submitEvent').addEventListener('click', async function () {
    const eventType = document.getElementById('eventType').value;
    const eventTitle = document.getElementById('eventTitle').value;
    const eventStartTime = document.getElementById('eventStartTime').value;
    const eventEndTime = document.getElementById('eventEndTime').value;

    // Make an API request to submit the event
    try {
        const response = await axios.post('/create', {
            type: eventType,
            title: eventTitle,
            start: eventStartTime,
            end: eventEndTime,
        });

        // Handle the API response (e.g., show a success message)
        if (response.data.status === 'OK') {
            alert('Event created successfully.');
            $('#eventModal').modal('hide');
        } else {
            alert('Failed to create the event. Please try again or correct the data.');
        }
    } catch (error) {
        console.log(error)
        alert('An error occurred while creating the event.');
    }
});



// Simulated API call to check user's permission to delete
async function checkUserPermissionToDelete() {
    // Replace with your API endpoint for checking user's permission
    return true;
    try {
        const response = await axios.get('/api/checkUserPermissionToDelete');
        return response.data.allowed;
    } catch (error) {
        return false; // Handle the error appropriately
    }
}
