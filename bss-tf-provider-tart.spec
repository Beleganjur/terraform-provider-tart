openapi: "3.0.1"
info:
  title: "Orchard Tart VM Orchestrator"
  version: "1.0.0"
  description: "API for lifecycle management of Tart virtual machines via Orchard."

servers:
  - url: "http://controller:8085/api"

paths:
  /vms:
    post:
      summary: "Create a new Tart VM"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                image:
                  type: string
              required: [name, image]
      responses:
        "201":
          description: "VM created"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Vm"
    get:
      summary: "List all Tart VMs"
      responses:
        "200":
          description: "A list of VMs"
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/Vm"

  /vms/{vm_id}:
    get:
      summary: "Get VM details"
      parameters:
        - name: vm_id
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: "Details for a Tart VM"
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Vm"
    delete:
      summary: "Delete a Tart VM"
      parameters:
        - name: vm_id
          in: path
          required: true
          schema:
            type: string
      responses:
        "204":
          description: "VM deleted, no body"

components:
  schemas:
    Vm:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        image:
          type: string
        status:
          type: string
          enum: [running, stopped, error]
